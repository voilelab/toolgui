//go:build js && wasm

package tgwasm

import (
	"log/slog"
	"sync"
	"syscall/js"

	"github.com/voilelab/toolgui/toolgui/tgframe"
	"github.com/voilelab/toolgui/toolgui/tgjson"
	"github.com/voilelab/toolgui/toolgui/tgutil"
)

// BridgeName is the global the bridge is installed on. The page calls
// globalThis.toolgui.<name>(...).
const BridgeName = "toolgui"

// ErrNoSession is returned when the page acts before calling start.
var ErrNoSession = tgutil.NewError("no session, call start first")

// ErrNoUpload is returned for a file the bridge never reserved, or reserved
// for a session that is over.
var ErrNoUpload = tgutil.NewError("no such upload")

// bridge is what the page talks to. It holds one session, because a tab is
// one user: there is nothing to key a session map by.
//
// Payloads cross as JSON strings, the same ones the web executor sends over
// its websocket, so both transports carry an identical wire format.
type bridge struct {
	app *tgframe.App

	// onPack is the callback the page registered for packs.
	onPack js.Value

	// lock guards the session and its state: a page func runs on its own
	// goroutine while calls arrive from the browser's.
	lock    sync.Mutex
	session *tgframe.Session
	state   *tgframe.State

	// uploads are the files reserved for the page and not yet handed back,
	// keyed by the name the store gave them. An upload is two calls -- one to
	// reserve, one to hand over -- because the writing in between is the
	// page's and cannot be awaited from here.
	uploads map[string]*tgframe.BrowserUpload
}

func newBridge(app *tgframe.App) *bridge {
	return &bridge{
		app:     app,
		onPack:  js.Undefined(),
		uploads: map[string]*tgframe.BrowserUpload{},
	}
}

// install publish the bridge on the global object. Everything on it returns
// at once: a call that blocks would hand control back to JavaScript before
// its work is done, so results come back as packs instead.
func (b *bridge) install() {
	js.Global().Set(BridgeName, map[string]any{
		"appConf":      js.FuncOf(b.jsAppConf),
		"onPack":       js.FuncOf(b.jsOnPack),
		"start":        js.FuncOf(b.jsStart),
		"update":       js.FuncOf(b.jsUpdate),
		"newUpload":    js.FuncOf(b.jsNewUpload),
		"uploadFile":   js.FuncOf(b.jsUploadFile),
		"cancelUpload": js.FuncOf(b.jsCancelUpload),
	})
}

// jsAppConf return the app config as JSON. It's the browser counterpart of
// GET /api/app.
func (b *bridge) jsAppConf(this js.Value, args []js.Value) any {
	bs, err := tgjson.Marshal(b.app.AppConf())
	if err != nil {
		// AppConf is plain data, so this cannot fail in practice.
		slog.Error("marshal app conf", "error", err)
		return ""
	}

	return string(bs)
}

// jsOnPack register the callback every pack is handed to, as one JSON string.
// It has to be called before start, or the first run's packs are lost.
func (b *bridge) jsOnPack(this js.Value, args []js.Value) any {
	if len(args) == 0 {
		return nil
	}

	b.onPack = args[0]
	return nil
}

// jsStart open a session on the given page and draw it once. Calling it again
// switches page: the previous session is closed and the new one starts from an
// empty state, the same as loading another page in the browser.
func (b *bridge) jsStart(this js.Value, args []js.Value) any {
	pageName := ""
	if len(args) > 0 {
		pageName = args[0].String()
	}

	b.lock.Lock()

	b.closeSession()

	state := tgframe.NewState()
	session, err := tgframe.NewSession(b.app, pageName, state, b.send)
	if err != nil {
		// The state never became the bridge's, so nothing else will let go of
		// what it holds.
		state.Destroy()
		b.lock.Unlock()

		// Only the page name can fail here, and a retry would fail the same
		// way.
		b.sendResult(&tgframe.ResultPack{
			Error: err.Error(),
			Fatal: true,
		})
		return nil
	}

	b.state = state
	b.session = session

	b.lock.Unlock()

	session.HandleEvent(&tgframe.EventEmpty{})
	return nil
}

// jsUpdate apply an event to the session and rerun the page. It's the browser
// counterpart of a message on the update websocket.
func (b *bridge) jsUpdate(this js.Value, args []js.Value) any {
	session := b.currentSession()
	if session == nil || len(args) == 0 {
		b.sendResult(&tgframe.ResultPack{Error: ErrNoSession.Error()})
		return nil
	}

	err := session.HandleRawEvent([]byte(args[0].String()))
	if err != nil {
		// HandleRawEvent already reported it to the page.
		slog.Error("handle event", "error", err)
	}

	return nil
}

// uploadSlot is where the page writes one upload, as jsNewUpload answers it.
// Dir names the directory from the origin private file system's root down, and
// Name is the file inside it, which is also what identifies the upload in the
// two calls that follow.
type uploadSlot struct {
	Dir   []string `json:"dir,omitempty"`
	Name  string   `json:"name,omitempty"`
	Error string   `json:"error,omitempty"`
}

// jsNewUpload reserve a file for the page to stream an upload into, and answer
// with where it is as JSON. It is the first half of the browser counterpart of
// POST /api/files.
//
// The page writes the file itself, with a writable stream, because that is the
// only way a picked file is copied without the whole of it passing through the
// tab's heap. Nothing here waits for that: the call returns a place to write
// and the page comes back with jsUploadFile when it is done.
func (b *bridge) jsNewUpload(this js.Value, args []js.Value) any {
	b.lock.Lock()
	defer b.lock.Unlock()

	if b.state == nil {
		return uploadFailed(ErrNoSession)
	}

	upload, err := b.state.NewBrowserUpload()
	if err != nil {
		return uploadFailed(err)
	}

	b.uploads[upload.Name()] = upload

	return marshalSlot(&uploadSlot{Dir: upload.Dir(), Name: upload.Name()})
}

// jsUploadFile take over a file the page has finished writing and store it
// under the component that asked for it. It is the second half of the browser
// counterpart of POST /api/files, and it answers with an error message, or an
// empty string on success, because the page waits for it.
//
// It takes the sync access handle rather than opening one: opening is a
// promise, and this arrives on the JavaScript callback stack where a promise
// cannot settle. The page opens it after closing its writable stream -- the
// two are exclusive holds on the same file, and the other order gets
// NoModificationAllowedError -- so by the time this runs there is nothing left
// to wait for.
//
// No bytes cross. What crosses is a component ID, the name the user's file had
// and the file already sitting in the origin private file system.
func (b *bridge) jsUploadFile(this js.Value, args []js.Value) any {
	// The lock is held across the handover: a start on another goroutine must
	// not replace the state this is storing into.
	b.lock.Lock()
	defer b.lock.Unlock()

	if len(args) < 4 {
		return ErrNoUpload.Error()
	}

	componentID, name, slot, handle := args[0].String(), args[1].String(),
		args[2].String(), args[3]

	if b.state == nil {
		closeHandle(handle)
		return ErrNoSession.Error()
	}

	upload := b.uploads[slot]
	if upload == nil {
		// A page switch while the page was writing takes the reservations of
		// the session before it, so this is where one arrives late.
		closeHandle(handle)
		return ErrNoUpload.Error()
	}

	delete(b.uploads, slot)

	file, err := upload.Take(name, handle)
	if err != nil {
		// Nothing half written becomes a file a page can read.
		upload.Discard()
		closeHandle(handle)
		return err.Error()
	}

	b.state.PutFile(componentID, file)

	return ""
}

// jsCancelUpload drop a reservation the page could not fill. An upload that
// failed halfway -- out of quota, cancelled, or a stream that broke -- leaves
// nothing in the origin private file system and never reaches the component.
func (b *bridge) jsCancelUpload(this js.Value, args []js.Value) any {
	b.lock.Lock()
	defer b.lock.Unlock()

	if len(args) == 0 {
		return nil
	}

	slot := args[0].String()

	upload := b.uploads[slot]
	if upload == nil {
		return nil
	}

	delete(b.uploads, slot)
	upload.Discard()

	return nil
}

// closeHandle lets go of a sync access handle the page opened and the bridge
// would not take. The page hands one over already open, so a refusal has to
// close it: the browser will not remove a file anything still holds one for,
// and the state's directory would outlive the session it belongs to.
//
// Closing twice is a no-op, which is what makes this safe after a handover
// that got far enough for the file to have closed it already.
func closeHandle(handle js.Value) {
	defer func() {
		if r := recover(); r != nil {
			slog.Error("close a handle the upload did not take", "error", r)
		}
	}()

	if !handle.Truthy() {
		return
	}

	handle.Call("close")
}

// uploadFailed is the answer to a reservation that could not be made.
func uploadFailed(err error) string {
	return marshalSlot(&uploadSlot{Error: err.Error()})
}

// marshalSlot renders a slot as JSON. The slot is plain data, so the only way
// this fails is a bug, and the page is told as much rather than left with an
// empty string it would have to guess about.
func marshalSlot(slot *uploadSlot) string {
	bs, err := tgjson.Marshal(slot)
	if err != nil {
		slog.Error("marshal an upload slot", "error", err)
		return `{"error":"cannot describe where to write the upload"}`
	}

	return string(bs)
}

// send push a pack to the page. [tgframe.Session] serializes the calls.
func (b *bridge) send(pack any) error {
	if b.onPack.IsUndefined() {
		return tgutil.NewError("no pack callback, call onPack first")
	}

	bs, err := tgjson.Marshal(pack)
	if err != nil {
		return tgutil.Errorf("%w", err)
	}

	b.onPack.Invoke(string(bs))
	return nil
}

// sendResult report an error the page has no other way to learn about. A
// failed send is only logged: there is nowhere left to report it to.
func (b *bridge) sendResult(pack *tgframe.ResultPack) {
	err := b.send(pack)
	if err != nil {
		slog.Error("send result pack", "error", err)
	}
}

func (b *bridge) currentSession() *tgframe.Session {
	b.lock.Lock()
	defer b.lock.Unlock()

	return b.session
}

// closeSession must be called with lock held.
//
// The state is destroyed, not merely dropped: what it holds is not all the
// garbage collector's to reclaim. In the browser its uploads are files in the
// origin private file system, with handles open on them, and a page switch
// that only let go of the pointer would leave every one of them behind for as
// long as the tab lived. [tgframe.Session.Close] does not do it -- the web
// executor destroys the state out of its own session map -- so it is done
// here, the way the desktop executor does.
func (b *bridge) closeSession() {
	if b.session == nil {
		return
	}

	b.session.Close()

	if b.state != nil {
		b.state.Destroy()
	}

	b.session = nil
	b.state = nil

	// The reservations go with the state that made them: destroying it takes
	// the whole directory, so there is nothing left for Discard to remove.
	clear(b.uploads)
}
