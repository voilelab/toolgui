//go:build js && wasm

package tgwasm

import (
	"log/slog"
	"syscall/js"

	"github.com/voilelab/toolgui/toolgui/tgframe"
	"github.com/voilelab/toolgui/toolgui/tgjson"
	"github.com/voilelab/toolgui/toolgui/tgutil"
)

// BridgeName is the global the bridge is installed on. The page calls
// globalThis.toolgui.<name>(...).
const BridgeName = "toolgui"

// ErrNoSession is returned when the page acts before calling start.
var ErrNoSession = tgframe.ErrNoSession

// ErrNoUpload is returned for an upload never reserved, or reserved by an
// ended session.
var ErrNoUpload = tgutil.NewError("no such upload")

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

	// host holds the session. Its uploads are reserved files not yet handed
	// back, by name: the page does the writing between the two calls.
	host *tgframe.SessionHost[*tgframe.BrowserUpload]

	// runs does the page-func part of start and update in call order, off
	// the JS call so a page func awaiting a promise doesn't deadlock.
	runs *serial
}

func newBridge(app *tgframe.App) *bridge {
	b := &bridge{
		app:     app,
		onPack:  js.Undefined(),
		onEvent: js.Undefined(),
		runs:    newSerial(),
	}
	b.host = tgframe.NewSessionHost[*tgframe.BrowserUpload](app, b.send)
	return b
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

	session, closeOld, err := b.host.Start(pageName, rawQuery)

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
	session := b.host.Session()
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
	var slot downloadSlot

	err := b.host.Do(func(state *tgframe.State, _ map[string]*tgframe.BrowserUpload) error {
		if len(args) == 0 {
			return ErrNoDownload
		}

		download := state.GetDownload(args[0].String())
		if download == nil {
			return ErrNoDownload
		}

		dir, name, err := download.BrowserLocation()
		if err != nil {
			return err
		}

		slot = downloadSlot{Dir: dir, Name: name}
		return nil
	})
	if err != nil {
		slot = downloadSlot{Error: err.Error()}
	}

	return marshalDownloadSlot(&slot)
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
	var slot uploadSlot

	err := b.host.Do(func(state *tgframe.State, uploads map[string]*tgframe.BrowserUpload) error {
		upload, err := state.NewBrowserUpload()
		if err != nil {
			return err
		}

		uploads[upload.Name()] = upload
		slot = uploadSlot{Dir: upload.Dir(), Name: upload.Name()}
		return nil
	})
	if err != nil {
		return uploadFailed(err)
	}

	return marshalSlot(&slot)
}

// jsUploadFile take over a file the page finished writing and store it under
// the component. Second half of POST /api/files; answers an error message, or
// "" on success.
//
// The page passes an open sync access handle (opened after closing its
// writable stream), since opening is a promise that can't settle on the JS
// callback stack. No bytes cross.
func (b *bridge) jsUploadFile(this js.Value, args []js.Value) any {
	if len(args) < 4 {
		return ErrNoUpload.Error()
	}

	componentID, name, slot, handle := args[0].String(), args[1].String(),
		args[2].String(), args[3]

	// Do holds the lock across the handover so a start can't swap the state.
	err := b.host.Do(func(state *tgframe.State, uploads map[string]*tgframe.BrowserUpload) error {
		upload := uploads[slot]
		if upload == nil {
			// A page switch while writing dropped the reservation.
			return ErrNoUpload
		}

		delete(uploads, slot)

		// Same check as POST /api/files.
		if !state.HasFileKey(componentID) {
			upload.Discard()
			return tgframe.ErrNotOnPage
		}

		file, err := upload.Take(name, handle)
		if err != nil {
			// Nothing half written becomes a file a page can read.
			upload.Discard()
			return err
		}

		state.PutFile(componentID, file)
		return nil
	})
	if err != nil {
		closeHandle(handle)
		return err.Error()
	}

	return ""
}

// jsCancelUpload drop a reservation the page could not fill (quota,
// cancelled, broken stream), leaving nothing behind.
func (b *bridge) jsCancelUpload(this js.Value, args []js.Value) any {
	if len(args) == 0 {
		return nil
	}

	slot := args[0].String()

	// No session means the reservation is gone already.
	_ = b.host.Do(func(_ *tgframe.State, uploads map[string]*tgframe.BrowserUpload) error {
		if upload := uploads[slot]; upload != nil {
			delete(uploads, slot)
			upload.Discard()
		}
		return nil
	})

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
