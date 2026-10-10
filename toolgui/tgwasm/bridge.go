//go:build js && wasm

package tgwasm

import (
	"log/slog"
	"sync"
	"sync/atomic"
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

// ErrNoUpload is returned for an upload never reserved, or reserved by an
// ended session.
var ErrNoUpload = tgutil.NewError("no such upload")

// errDetached is returned to sends of a session replaced by a later start.
var errDetached = tgutil.NewError("session replaced by a later start")

// ErrNoDownload is returned for an unknown or replaced download token.
var ErrNoDownload = tgutil.NewError("no such download")

// bridge is what the page talks to. It holds one session, since a tab is one
// user. Payloads are the same JSON strings the web executor's websocket uses.
type bridge struct {
	app *tgframe.App

	// onPack is the callback the page registered for packs.
	onPack js.Value

	// onEvent is the callback the page registered for [Emit].
	onEvent js.Value

	// lock guards session and state, shared by page funcs and JS calls.
	lock    sync.Mutex
	session *tgframe.Session
	state   *tgframe.State

	// detached cuts off the session's sends once a start replaces it, since
	// its close may wait behind queued runs.
	detached *atomic.Bool

	// uploads are reserved files not yet handed back, by name. Reserve and
	// hand-over are separate calls since the page does the writing between.
	uploads map[string]*tgframe.BrowserUpload

	// runs does the page-func part of start and update in call order, off
	// the JS call so a page func awaiting a promise doesn't deadlock.
	runs *serial
}

func newBridge(app *tgframe.App) *bridge {
	return &bridge{
		app:     app,
		onPack:  js.Undefined(),
		onEvent: js.Undefined(),
		uploads: map[string]*tgframe.BrowserUpload{},
		runs:    newSerial(),
	}
}

// install publish the bridge on the global object. Every call returns at
// once; results come back as packs.
func (b *bridge) install() {
	js.Global().Set(BridgeName, map[string]any{
		"appConf":      js.FuncOf(b.jsAppConf),
		"onPack":       js.FuncOf(b.jsOnPack),
		"onEvent":      js.FuncOf(b.jsOnEvent),
		"start":        js.FuncOf(b.jsStart),
		"update":       js.FuncOf(b.jsUpdate),
		"downloadFile": js.FuncOf(b.jsDownloadFile),
		"newUpload":    js.FuncOf(b.jsNewUpload),
		"uploadFile":   js.FuncOf(b.jsUploadFile),
		"cancelUpload": js.FuncOf(b.jsCancelUpload),
	})

	emitter = b
	bridgeRunning.Store(true)
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

// jsOnPack register the callback that receives each pack as JSON. Call it
// before start, or the first run's packs are lost.
func (b *bridge) jsOnPack(this js.Value, args []js.Value) any {
	if len(args) == 0 {
		return nil
	}

	b.onPack = args[0]
	return nil
}

// jsStart open a session on the given page and draw it once. Calling it again
// switches page, closing the old session and starting from an empty state.
//
// The second argument is the page query, `group=a` of `#/detail?group=a`.
func (b *bridge) jsStart(this js.Value, args []js.Value) any {
	pageName := ""
	if len(args) > 0 {
		pageName = args[0].String()
	}

	rawQuery := ""
	if len(args) > 1 && args[1].Type() == js.TypeString {
		rawQuery = args[1].String()
	}

	detached := new(atomic.Bool)
	send := func(pack any) error {
		if detached.Load() {
			return errDetached
		}
		return b.send(pack)
	}

	state := tgframe.NewState()
	query, err := tgframe.ParseQuery(rawQuery)
	var session *tgframe.Session
	if err == nil {
		session, err = tgframe.NewSession(b.app, pageName, query, state, send)
	}

	b.lock.Lock()
	closeOld := b.detachSession()
	if err == nil {
		b.state = state
		b.session = session
		b.detached = detached
	}
	b.lock.Unlock()

	if err != nil {
		// The state never reached the bridge, so free it here.
		state.Destroy()
	}

	// Queue the fatal result too, or a queued old run's ready pack clears it.
	b.runs.do(func() {
		closeOld()
		if err != nil {
			// Bad page name or query: a retry would fail too.
			b.sendResult(&tgframe.ResultPack{
				Error: err.Error(),
				Fatal: true,
			})
			return
		}
		session.HandleEvent(&tgframe.EventEmpty{})
	})

	return nil
}

// jsUpdate apply an event to the session and rerun the page, like a message
// on the update websocket.
func (b *bridge) jsUpdate(this js.Value, args []js.Value) any {
	session := b.currentSession()
	if session == nil || len(args) == 0 {
		b.sendResult(&tgframe.ResultPack{Error: ErrNoSession.Error()})
		return nil
	}

	raw := []byte(args[0].String())
	b.runs.do(func() {
		err := session.HandleRawEvent(raw)
		if err != nil {
			// HandleRawEvent already reported it to the page.
			slog.Error("handle event", "error", err)
		}
	})

	return nil
}

// downloadSlot is jsDownloadFile's answer: the OPFS directory path and file
// name, or an error.
type downloadSlot struct {
	Dir   []string `json:"dir,omitempty"`
	Name  string   `json:"name,omitempty"`
	Error string   `json:"error,omitempty"`
}

// jsDownloadFile answer where the file behind a download token is, as JSON,
// like GET /api/files.
//
// No bytes cross: the page reads the file with getFile, a disk-backed blob.
// Only this state's tokens are found.
func (b *bridge) jsDownloadFile(this js.Value, args []js.Value) any {
	b.lock.Lock()
	defer b.lock.Unlock()

	if b.state == nil {
		return marshalDownloadSlot(&downloadSlot{Error: ErrNoSession.Error()})
	}

	if len(args) == 0 {
		return marshalDownloadSlot(&downloadSlot{Error: ErrNoDownload.Error()})
	}

	download := b.state.GetDownload(args[0].String())
	if download == nil {
		return marshalDownloadSlot(&downloadSlot{Error: ErrNoDownload.Error()})
	}

	dir, name, err := download.BrowserLocation()
	if err != nil {
		return marshalDownloadSlot(&downloadSlot{Error: err.Error()})
	}

	return marshalDownloadSlot(&downloadSlot{Dir: dir, Name: name})
}

// uploadSlot is jsNewUpload's answer: the OPFS directory path and file name
// to write to. Name also identifies the upload in later calls.
type uploadSlot struct {
	Dir   []string `json:"dir,omitempty"`
	Name  string   `json:"name,omitempty"`
	Error string   `json:"error,omitempty"`
}

// jsNewUpload reserve a file for the page to stream an upload into, and answer
// where it is as JSON. First half of POST /api/files.
//
// The page writes it with a writable stream, keeping the file out of the
// heap, then calls jsUploadFile.
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

// jsUploadFile take over a file the page finished writing and store it under
// the component. Second half of POST /api/files; answers an error message, or
// "" on success.
//
// The page passes an open sync access handle (opened after closing its
// writable stream), since opening is a promise that can't settle on the JS
// callback stack. No bytes cross.
func (b *bridge) jsUploadFile(this js.Value, args []js.Value) any {
	// Held across the handover so a start can't swap the state.
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
		// A page switch while writing dropped the reservation.
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

// jsCancelUpload drop a reservation the page could not fill (quota,
// cancelled, broken stream), leaving nothing behind.
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

// closeHandle closes a sync access handle the bridge refused, or the state's
// directory can't be removed. Closing twice is a no-op.
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

// marshalDownloadSlot renders a download slot as JSON.
func marshalDownloadSlot(slot *downloadSlot) string {
	bs, err := tgjson.Marshal(slot)
	if err != nil {
		slog.Error("marshal a download slot", "error", err)
		return `{"error":"cannot describe where the download is"}`
	}

	return string(bs)
}

// marshalSlot renders a slot as JSON. Failure is a bug, reported as an error.
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

// sendResult report an error to the page. A failed send is only logged.
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

// detachSession takes the session off the bridge and returns its closer,
// which waits for the page func and so must not run on a JS call. It must be
// called with lock held.
//
// The state is destroyed, not just dropped: its uploads are OPFS files with
// open handles that GC won't reclaim.
func (b *bridge) detachSession() (closeSession func()) {
	session, state := b.session, b.state
	b.session = nil
	b.state = nil

	if b.detached != nil {
		b.detached.Store(true)
		b.detached = nil
	}

	// Reservations go with their state; a late handover finds no slot and
	// closes its handle.
	clear(b.uploads)

	return func() {
		if session != nil {
			session.Close()
		}
		if state != nil {
			state.Destroy()
		}
	}
}
