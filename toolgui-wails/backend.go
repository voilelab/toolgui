package tgwails

import (
	"bytes"
	"context"
	"encoding/base64"
	"sync"
	"uuid"

	"github.com/voilelab/toolgui/toolgui/tgframe"
	"github.com/voilelab/toolgui/toolgui/tgjson"
	"github.com/voilelab/toolgui/toolgui/tgutil"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// PackEventName is the Wails event every pack is emitted on; one name keeps
// packs in order.
const PackEventName = "toolgui:pack"

// ErrNoSession is returned when the frontend acts before calling Start.
var ErrNoSession = tgframe.ErrNoSession

// ErrNoUpload is returned for an unknown upload id, e.g. a chunk sent after
// the session was replaced.
var ErrNoUpload = tgutil.NewError("no such upload")

// ToolGUI is the struct Wails binds. Its exported methods reach the frontend
// as window.go.tgwails.ToolGUI.<Method>, each returning a promise. Payloads
// are the same JSON strings the web executor's websocket uses.
type ToolGUI struct {
	app *tgframe.App

	// emit hands one pack to the frontend. Set on window start (nil in
	// tests).
	emit func(packJSON string)

	// host holds the session; its uploads are files still arriving, by
	// upload id.
	host *tgframe.SessionHost[*tgframe.File]

	// menuLock guards the menu pick queue, so queueing never waits on a
	// run.
	menuLock  sync.Mutex
	menuQueue []string
	menuBusy  bool
}

// NewToolGUI return the bound struct serving app.
func NewToolGUI(app *tgframe.App) *ToolGUI {
	t := &ToolGUI{app: app}
	t.host = tgframe.NewSessionHost[*tgframe.File](app, t.send)
	return t
}

// start is wired to [options.App.OnStartup], unexported so the frontend
// can't call it.
func (t *ToolGUI) start(ctx context.Context) {
	t.emit = func(packJSON string) {
		runtime.EventsEmit(ctx, PackEventName, packJSON)
	}
}

// shutdown is wired to [options.App.OnShutdown].
func (t *ToolGUI) shutdown(ctx context.Context) {
	t.host.Close()()
}

// AppConf return the app config as JSON, like GET /api/app.
//
// The menu is dropped: it goes to the native menubar (see [Executor.Run]).
func (t *ToolGUI) AppConf() (string, error) {
	conf := *t.app.AppConf()
	conf.Menu = nil

	bs, err := tgjson.Marshal(&conf)
	if err != nil {
		return "", tgutil.Errorf("%w", err)
	}

	return string(bs), nil
}

// clickMenu applies a native menubar click on id, the same event the web
// menubar sends. Unexported so the frontend can't fire it. A click before
// Start is ignored.
func (t *ToolGUI) clickMenu(id string) {
	session := t.host.Session()
	if session == nil {
		return
	}

	session.HandleEvent(&tgframe.EventClick{ID: id})
}

// queueMenuClick queues a native menubar pick and ensures one goroutine
// drains the queue in order.
//
// Running inline would freeze the Windows message loop; a goroutine per pick
// could reorder picks. (macOS/Linux may already reorder before this.)
func (t *ToolGUI) queueMenuClick(id string) {
	t.menuLock.Lock()
	t.menuQueue = append(t.menuQueue, id)

	if t.menuBusy {
		t.menuLock.Unlock()
		return
	}

	t.menuBusy = true
	t.menuLock.Unlock()

	go t.drainMenuClicks()
}

// drainMenuClicks applies queued picks oldest first until empty. menuBusy,
// guarded by menuLock, ensures only one runs at a time.
func (t *ToolGUI) drainMenuClicks() {
	for {
		t.menuLock.Lock()
		if len(t.menuQueue) == 0 {
			t.menuQueue = nil
			t.menuBusy = false
			t.menuLock.Unlock()

			return
		}

		id := t.menuQueue[0]
		t.menuQueue = t.menuQueue[1:]
		t.menuLock.Unlock()

		t.clickMenu(id)
	}
}

// Start open a session on pageName and run the page once. Calling it again
// switches page, starting from an empty state. query (`group=a`) is read as
// [tgframe.Params.Query].
func (t *ToolGUI) Start(pageName, query string) error {
	session, closeOld, err := t.host.Start(pageName, query)

	// Always close: the frontend has already switched page.
	closeOld()

	if err != nil {
		return tgutil.Errorf("%w", err)
	}

	// Draw the page for the first time.
	session.HandleEvent(&tgframe.EventEmpty{})
	return nil
}

// Update apply a frontend event to the session and rerun the page, like the
// update websocket.
func (t *ToolGUI) Update(eventJSON string) error {
	session := t.host.Session()
	if session == nil {
		return ErrNoSession
	}

	err := session.HandleRawEvent([]byte(eventJSON))
	if err != nil {
		return tgutil.Errorf("%w", err)
	}

	return nil
}

// UploadFileStart opens a file for an upload and returns the id its chunks
// carry. Chunked, unlike POST /api/files, since the bridge only takes strings.
func (t *ToolGUI) UploadFileStart(name string) (string, error) {
	var uploadID string

	err := t.host.Do(func(state *tgframe.State, uploads map[string]*tgframe.File) error {
		file, err := state.NewFile(name)
		if err != nil {
			return tgutil.Errorf("%w", err)
		}

		uploadID = uuid.New().String()
		uploads[uploadID] = file
		return nil
	})

	return uploadID, err
}

// UploadFileChunk appends one base64 encoded chunk to the upload.
func (t *ToolGUI) UploadFileChunk(uploadID, dataBase64 string) error {
	bs, err := base64.StdEncoding.DecodeString(dataBase64)
	if err != nil {
		return tgutil.Errorf("%w", err)
	}

	return t.host.Do(func(_ *tgframe.State, uploads map[string]*tgframe.File) error {
		file, ok := uploads[uploadID]
		if !ok {
			return ErrNoUpload
		}

		if err := file.Append(bytes.NewReader(bs)); err != nil {
			return tgutil.Errorf("%w", err)
		}

		return nil
	})
}

// UploadFileFinish hands the finished upload to the component. The page only
// sees it from here, so a second pick replaces the first.
func (t *ToolGUI) UploadFileFinish(componentID, uploadID string) error {
	return t.host.Do(func(state *tgframe.State, uploads map[string]*tgframe.File) error {
		file, ok := uploads[uploadID]
		if !ok {
			return ErrNoUpload
		}

		delete(uploads, uploadID)

		// Same check as POST /api/files: a key no component reads is never
		// released. The dropped file goes with the state.
		if !state.HasFileKey(componentID) {
			return tgframe.ErrNotOnPage
		}

		state.PutFile(componentID, file)
		return nil
	})
}

// send push a pack to the frontend. Calls are serialized and Wails keeps
// their order.
func (t *ToolGUI) send(pack any) error {
	if t.emit == nil {
		return tgutil.NewError("wails runtime is not ready")
	}

	bs, err := tgjson.Marshal(pack)
	if err != nil {
		return tgutil.Errorf("%w", err)
	}

	t.emit(string(bs))
	return nil
}
