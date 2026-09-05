package tgwails

import (
	"encoding/base64"
	"encoding/json"
	"sync"

	"github.com/voilelab/toolgui/toolgui/tgframe"
	"github.com/voilelab/toolgui/toolgui/tgutil"

	"github.com/wailsapp/wails"
)

// PackEventName is the Wails event every pack is emitted on. A single event
// name keeps create/update/delete/result in the order the page produced them.
const PackEventName = "toolgui:pack"

// ErrNoSession is returned when the frontend acts before calling Start.
var ErrNoSession = tgutil.NewError("no session, call Start first")

// ToolGUI is the struct Wails binds. Its exported methods reach the frontend
// as window.backend.ToolGUI.<Method>, each returning a promise.
//
// Wails converts a JSON argument straight to the parameter type, which rules
// out structs, so every payload crosses as a string.
type ToolGUI struct {
	app *tgframe.App

	runtime *wails.Runtime

	// lock guards the session and the state it runs on. Bound methods are
	// called from the IPC goroutine, WailsShutdown from another.
	lock    sync.Mutex
	session *tgframe.Session
	state   *tgframe.State
}

// NewToolGUI return the bound struct serving app.
func NewToolGUI(app *tgframe.App) *ToolGUI {
	return &ToolGUI{app: app}
}

// WailsInit is called by Wails before the frontend starts.
func (t *ToolGUI) WailsInit(runtime *wails.Runtime) error {
	t.runtime = runtime
	return nil
}

// WailsShutdown is called by Wails when the window closes.
func (t *ToolGUI) WailsShutdown() {
	t.lock.Lock()
	defer t.lock.Unlock()

	t.closeSession()
}

// AppConf return the app config as JSON. It's the desktop counterpart of
// GET /api/app.
func (t *ToolGUI) AppConf() (string, error) {
	bs, err := json.Marshal(t.app.AppConf())
	if err != nil {
		return "", tgutil.Errorf("%w", err)
	}

	return string(bs), nil
}

// Start open a session on pageName and run the page once. Calling it again
// switches page: the previous session is closed and the new one starts from
// an empty state, the same as loading another page in the browser.
func (t *ToolGUI) Start(pageName string) error {
	t.lock.Lock()

	t.closeSession()

	state := tgframe.NewState()
	session, err := tgframe.NewSession(t.app, pageName, state, t.send)
	if err != nil {
		t.lock.Unlock()
		return tgutil.Errorf("%w", err)
	}

	t.state = state
	t.session = session

	t.lock.Unlock()

	// Draw the page for the first time.
	session.HandleEvent(&tgframe.EventEmpty{})
	return nil
}

// Update apply a frontend event to the session and rerun the page. It's the
// desktop counterpart of the update websocket.
func (t *ToolGUI) Update(eventJSON string) error {
	session := t.currentSession()
	if session == nil {
		return ErrNoSession
	}

	err := session.HandleRawEvent([]byte(eventJSON))
	if err != nil {
		return tgutil.Errorf("%w", err)
	}

	return nil
}

// UploadFile store a base64 encoded file in the session state. It's the
// desktop counterpart of POST /api/files.
func (t *ToolGUI) UploadFile(name string, dataBase64 string) error {
	t.lock.Lock()
	state := t.state
	t.lock.Unlock()

	if state == nil {
		return ErrNoSession
	}

	bs, err := base64.StdEncoding.DecodeString(dataBase64)
	if err != nil {
		return tgutil.Errorf("%w", err)
	}

	state.SetFile(name, bs)
	return nil
}

// send push a pack to the frontend. [tgframe.Session] serializes the calls,
// and the Wails event queue preserves their order.
func (t *ToolGUI) send(pack any) error {
	if t.runtime == nil {
		return tgutil.NewError("wails runtime is not ready")
	}

	bs, err := json.Marshal(pack)
	if err != nil {
		return tgutil.Errorf("%w", err)
	}

	t.runtime.Events.Emit(PackEventName, string(bs))
	return nil
}

func (t *ToolGUI) currentSession() *tgframe.Session {
	t.lock.Lock()
	defer t.lock.Unlock()

	return t.session
}

// closeSession must be called with lock held.
func (t *ToolGUI) closeSession() {
	if t.session == nil {
		return
	}

	t.session.Close()
	t.session = nil
	t.state = nil
}
