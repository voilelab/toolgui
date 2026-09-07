//go:build js && wasm

package tgwasm

import (
	"encoding/base64"
	"encoding/json"
	"log/slog"
	"sync"
	"syscall/js"

	"github.com/voilelab/toolgui/toolgui/tgframe"
	"github.com/voilelab/toolgui/toolgui/tgutil"
)

// BridgeName is the global the bridge is installed on. The page calls
// globalThis.toolgui.<name>(...).
const BridgeName = "toolgui"

// ErrNoSession is returned when the page acts before calling start.
var ErrNoSession = tgutil.NewError("no session, call start first")

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
}

func newBridge(app *tgframe.App) *bridge {
	return &bridge{app: app, onPack: js.Undefined()}
}

// install publish the bridge on the global object. Everything on it returns
// at once: a call that blocks would hand control back to JavaScript before
// its work is done, so results come back as packs instead.
func (b *bridge) install() {
	js.Global().Set(BridgeName, map[string]any{
		"appConf":    js.FuncOf(b.jsAppConf),
		"onPack":     js.FuncOf(b.jsOnPack),
		"start":      js.FuncOf(b.jsStart),
		"update":     js.FuncOf(b.jsUpdate),
		"uploadFile": js.FuncOf(b.jsUploadFile),
	})
}

// jsAppConf return the app config as JSON. It's the browser counterpart of
// GET /api/app.
func (b *bridge) jsAppConf(this js.Value, args []js.Value) any {
	bs, err := json.Marshal(b.app.AppConf())
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

// jsUploadFile store a base64 encoded file in the session state. It's the
// browser counterpart of POST /api/files. It answers with an error message, or
// an empty string on success, because the page waits for it.
func (b *bridge) jsUploadFile(this js.Value, args []js.Value) any {
	b.lock.Lock()
	state := b.state
	b.lock.Unlock()

	if state == nil || len(args) < 2 {
		return ErrNoSession.Error()
	}

	bs, err := base64.StdEncoding.DecodeString(args[1].String())
	if err != nil {
		return err.Error()
	}

	state.SetFile(args[0].String(), bs)
	return ""
}

// send push a pack to the page. [tgframe.Session] serializes the calls.
func (b *bridge) send(pack any) error {
	if b.onPack.IsUndefined() {
		return tgutil.NewError("no pack callback, call onPack first")
	}

	bs, err := json.Marshal(pack)
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
func (b *bridge) closeSession() {
	if b.session == nil {
		return
	}

	b.session.Close()
	b.session = nil
	b.state = nil
}
