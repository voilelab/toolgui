package tgwails

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"sync"
	"uuid"

	"github.com/voilelab/toolgui/toolgui/tgframe"
	"github.com/voilelab/toolgui/toolgui/tgutil"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// PackEventName is the Wails event every pack is emitted on. A single event
// name keeps create/update/delete/result in the order the page produced them.
const PackEventName = "toolgui:pack"

// ErrNoSession is returned when the frontend acts before calling Start.
var ErrNoSession = tgutil.NewError("no session, call Start first")

// ErrNoUpload is returned for an upload id the session doesn't know, which is
// what a chunk sent after the session was replaced looks like.
var ErrNoUpload = tgutil.NewError("no such upload")

// ToolGUI is the struct Wails binds. Its exported methods reach the frontend
// as window.go.tgwails.ToolGUI.<Method>, each returning a promise.
//
// Payloads cross as JSON strings, the same ones the web executor sends over
// its websocket, so both transports carry an identical wire format.
type ToolGUI struct {
	app *tgframe.App

	// emit hands one pack to the frontend. It's wired up when the window
	// starts, which is also what lets the tests run without one.
	emit func(packJSON string)

	// lock guards the session and the state it runs on: bound methods are
	// called from the frontend's goroutine, shutdown from another.
	lock    sync.Mutex
	session *tgframe.Session
	state   *tgframe.State

	// uploads are the files still arriving, by upload id. The lock guards
	// them too: a chunk that lands while the session is being replaced must
	// not write to the state that is going away.
	uploads map[string]*tgframe.File
}

// NewToolGUI return the bound struct serving app.
func NewToolGUI(app *tgframe.App) *ToolGUI {
	return &ToolGUI{app: app, uploads: make(map[string]*tgframe.File)}
}

// start is wired to [options.App.OnStartup] rather than being a bound method,
// so the frontend never sees it. The context it gets is what the Wails
// runtime needs to reach this window.
func (t *ToolGUI) start(ctx context.Context) {
	t.emit = func(packJSON string) {
		runtime.EventsEmit(ctx, PackEventName, packJSON)
	}
}

// shutdown is wired to [options.App.OnShutdown].
func (t *ToolGUI) shutdown(ctx context.Context) {
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
	t.uploads = make(map[string]*tgframe.File)

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

// UploadFileStart opens a file for an upload and returns the id the chunks
// after it carry. It's the desktop half of POST /api/files, which the bridge
// can't do in one call: it takes strings, so a whole file would sit in memory
// as a blob, as base64 and as bytes at once.
func (t *ToolGUI) UploadFileStart(name string) (string, error) {
	t.lock.Lock()
	defer t.lock.Unlock()

	if t.state == nil {
		return "", ErrNoSession
	}

	file, err := t.state.NewFile(name)
	if err != nil {
		return "", tgutil.Errorf("%w", err)
	}

	uploadID := uuid.New().String()
	t.uploads[uploadID] = file

	return uploadID, nil
}

// UploadFileChunk appends one base64 encoded chunk to the upload.
func (t *ToolGUI) UploadFileChunk(uploadID, dataBase64 string) error {
	bs, err := base64.StdEncoding.DecodeString(dataBase64)
	if err != nil {
		return tgutil.Errorf("%w", err)
	}

	t.lock.Lock()
	defer t.lock.Unlock()

	file, ok := t.uploads[uploadID]
	if !ok {
		return ErrNoUpload
	}

	if err := file.Append(bytes.NewReader(bs)); err != nil {
		return tgutil.Errorf("%w", err)
	}

	return nil
}

// UploadFileFinish hands the finished upload to the component that asked for
// it. Until then the page doesn't see it, so a second pick on the same
// component replaces the first rather than being spliced into it.
func (t *ToolGUI) UploadFileFinish(componentID, uploadID string) error {
	t.lock.Lock()
	defer t.lock.Unlock()

	if t.state == nil {
		return ErrNoSession
	}

	file, ok := t.uploads[uploadID]
	if !ok {
		return ErrNoUpload
	}

	delete(t.uploads, uploadID)
	t.state.PutFile(componentID, file)

	return nil
}

// send push a pack to the frontend. [tgframe.Session] serializes the calls,
// and the Wails event queue preserves their order.
func (t *ToolGUI) send(pack any) error {
	if t.emit == nil {
		return tgutil.NewError("wails runtime is not ready")
	}

	bs, err := json.Marshal(pack)
	if err != nil {
		return tgutil.Errorf("%w", err)
	}

	t.emit(string(bs))
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

	// The state owns the files uploaded to it, the ones still arriving
	// included, and nothing else can reach them once the session is gone.
	t.state.Destroy()
	t.state = nil
	t.uploads = make(map[string]*tgframe.File)
}
