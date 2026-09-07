package tgexec

import (
	"encoding/json"
	"io"
	"io/fs"
	"log/slog"
	"strings"
	"sync"
	"time"

	"net/http"

	toolguiweb "github.com/voilelab/toolgui/toolgui-web"
	"github.com/voilelab/toolgui/toolgui/tgframe"
	"github.com/voilelab/toolgui/toolgui/tgutil"

	"golang.org/x/net/websocket"
)

// TODO: Let it be configurable

// MaxUploadSize limit the size of file uploading form.
const MaxUploadSize int64 = 1024 * 1024 * 1024

// ErrUpdateInterrupt is raise at panic when current state is going to interrupt
//
// Deprecated: use [tgframe.ErrUpdateInterrupt].
var ErrUpdateInterrupt = tgframe.ErrUpdateInterrupt

// FIXME: use errors.Is or errors.As
const forceClosedByRemoteStr = "An existing connection was forcibly closed by the remote host."

// WebExecutor is a web ui executor for ToolGUI.
type WebExecutor struct {
	rootAssets map[string][]byte

	stateMap tgutil.UUIDMap[tgframe.State]

	app *tgframe.App

	// confMu guards manifest and assets, which the app may set at any time,
	// including while handlers are already serving requests.
	confMu sync.RWMutex

	// manifest is nil until the app sets one, and nil serves the default.
	manifest *Manifest

	// assets is nil until the app sets one.
	assets fs.FS
}

type stateIDPack struct {
	StateID string `json:"state_id"`
}

// NewWebExecutor return a WebExecutor.
func NewWebExecutor(app *tgframe.App) *WebExecutor {
	return &WebExecutor{
		rootAssets: toolguiweb.GetRootAssets(),

		stateMap: tgutil.NewUUIDMap(
			tgframe.NewState, func(t *tgframe.State) { t.Destroy() },
			5*time.Minute),

		app: app,
	}
}

// defaultManifest returns the manifest served when the app sets none. The app
// title names it, so an app that sets a title doesn't repeat it here.
func (e *WebExecutor) defaultManifest() *Manifest {
	manifest := DefaultManifest()

	if title := e.app.AppConf().Title; title != "" {
		manifest.Name = title
		manifest.ShortName = title
	}

	return manifest
}

// SetManifest sets the web app manifest served at /manifest.json. A nil
// manifest goes back to the default, which [tgframe.App.SetTitle] names.
//
//	e.SetManifest(&tgexec.Manifest{
//		Name:      "My Tool",
//		ShortName: "My Tool",
//		Display:   "standalone",
//	})
func (e *WebExecutor) SetManifest(manifest *Manifest) {
	e.confMu.Lock()
	defer e.confMu.Unlock()

	e.manifest = manifest
}

// SetAssets serves the files at the root of fsys under /assets/, so an app can
// hand the browser files of its own: a manifest icon, an image a page links
// to. A nil fsys serves none.
//
//	//go:embed assets
//	var assets embed.FS
//
//	// assets/icon.png is served at /assets/icon.png.
//	sub, _ := fs.Sub(assets, "assets")
//	e.SetAssets(sub)
func (e *WebExecutor) SetAssets(fsys fs.FS) {
	e.confMu.Lock()
	defer e.confMu.Unlock()

	e.assets = fsys
}

// Destory release all resource.
func (e *WebExecutor) Destroy() {
	e.stateMap.Destroy()
}

func (e *WebExecutor) handleUpdate(ws *websocket.Conn) {
	pageName := ws.Request().PathValue("name")
	if !e.app.HasPage(pageName) {
		websocket.JSON.Send(ws, &tgframe.ResultPack{
			Error:   "page not found",
			Success: false,
			Fatal:   true,
		})
		slog.Error("page not found", "page", pageName)
		return
	}

	var pack stateIDPack
	err := websocket.JSON.Receive(ws, &pack)
	if err != nil {
		websocket.JSON.Send(ws, &tgframe.ResultPack{
			Error:   err.Error(),
			Success: false,
		})
		slog.Error("state id", "error", err)
		return
	}

	var stateID string
	stateID = pack.StateID

	state, alive := e.stateMap.Get(stateID)
	if state == nil {
		stateID = e.stateMap.New()
		state, _ = e.stateMap.Get(stateID)
		websocket.JSON.Send(ws, stateIDPack{
			StateID: stateID,
		})
	} else {
		if alive {
			websocket.JSON.Send(ws, &tgframe.ResultPack{
				Error:   "state id already alive",
				Success: false,
			})
			slog.Error("state id already alive", "state_id", stateID)
			return
		}

		e.stateMap.SetAlive(stateID, true)
	}

	session, err := tgframe.NewSession(e.app, pageName, state,
		func(pack any) error { return websocket.JSON.Send(ws, pack) })
	if err != nil {
		// NewSession only fails on the page name, so a retry would fail the
		// same way.
		websocket.JSON.Send(ws, &tgframe.ResultPack{
			Error:   err.Error(),
			Success: false,
			Fatal:   true,
		})
		slog.Error("new session", "error", err)
		return
	}

	for {
		var bs []byte
		err := websocket.Message.Receive(ws, &bs)
		if err != nil {
			if err == io.EOF || strings.Contains(err.Error(), forceClosedByRemoteStr) {
				// Connection closed

				session.Close()
				e.stateMap.SetAlive(stateID, false)
				break
			}

			websocket.JSON.Send(ws, &tgframe.ResultPack{
				Error:   err.Error(),
				Success: false,
			})
			slog.Error("state value change", "error", err)
			continue
		}

		err = session.HandleRawEvent(bs)
		if err != nil {
			slog.Error("handle event", "error", err)
		}
	}
}

func (e *WebExecutor) handleUpload(w http.ResponseWriter, req *http.Request) {
	stateID := req.Header.Get("STATE_ID")
	state, alive := e.stateMap.Get(stateID)
	if state == nil || !alive {
		http.Error(w, "State ID is invalid or not alive", http.StatusForbidden)
		return
	}

	err := req.ParseMultipartForm(MaxUploadSize)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.Error("Parse form", "error", err)
		return
	}

	file, handler, err := req.FormFile("file")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.Error("Get formfile", "error", err)
		return
	}
	defer file.Close()

	bs, err := io.ReadAll(file)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.Error("Open file", "error", err)
		return
	}

	// TODO: Remove old file
	state.SetFile(handler.Filename, bs)
}

func (e *WebExecutor) handlePage(resp http.ResponseWriter, req *http.Request) {
	pageName := req.PathValue("name")
	body, isRootAssets := e.rootAssets[pageName]
	if isRootAssets {
		resp.Write(body)
		return
	}

	resp.Write([]byte(toolguiweb.IndexBody))
}

func (e *WebExecutor) handleAssets(resp http.ResponseWriter, req *http.Request) {
	pageName := req.PathValue("name")
	body, isRootAssets := e.rootAssets[pageName]
	if !isRootAssets {
		resp.WriteHeader(http.StatusNotFound)
		return
	}

	resp.Write(body)
}

// handleAsset serves a file the app gave [WebExecutor.SetAssets]. Reading the
// fs per request, rather than at mux time, frees the app to set it whenever.
func (e *WebExecutor) handleAsset(resp http.ResponseWriter, req *http.Request) {
	e.confMu.RLock()
	assets := e.assets
	e.confMu.RUnlock()

	if assets == nil {
		http.NotFound(resp, req)
		return
	}

	http.FileServerFS(assets).ServeHTTP(resp, req)
}

func (e *WebExecutor) handleIndex(resp http.ResponseWriter, req *http.Request) {
	resp.Write([]byte(toolguiweb.IndexBody))
}

func (e *WebExecutor) handleManifest(resp http.ResponseWriter, req *http.Request) {
	e.confMu.RLock()
	manifest := e.manifest
	e.confMu.RUnlock()

	if manifest == nil {
		manifest = e.defaultManifest()
	}

	bs, err := json.Marshal(manifest)
	if err != nil {
		resp.WriteHeader(http.StatusInternalServerError)
		slog.Error("marshal manifest", "error", err)
		return
	}

	resp.Header().Set("Content-Type", "application/manifest+json")
	resp.Write(bs)
}

func (e *WebExecutor) handleHealth(resp http.ResponseWriter, req *http.Request) {
	resp.WriteHeader(http.StatusOK)
}

func (e *WebExecutor) handleAppConf(resp http.ResponseWriter, req *http.Request) {
	bs, err := json.Marshal(e.app.AppConf())
	if err != nil {
		resp.WriteHeader(http.StatusInternalServerError)
		return
	}

	resp.Write(bs)
}

// Mux return a http mux to handle whole app.
//
//	mux, _ := e.Mux()
//	http.ListenAndServe(":8080", mux)
func (e *WebExecutor) Mux() (*http.ServeMux, error) {
	mux := http.NewServeMux()

	// More specific than the page patterns below, so it wins over them.
	mux.HandleFunc("GET /manifest.json", e.handleManifest)

	if e.app.AppConf().HashPageNameMode {
		mux.HandleFunc("GET /{name}", e.handleAssets)
		mux.HandleFunc("GET /", e.handleIndex)
	} else {
		mux.HandleFunc("GET /{name}", e.handlePage)
		if firstPage, ok := e.app.FirstPage(); ok {
			mux.Handle("GET /", http.RedirectHandler("/"+firstPage,
				http.StatusTemporaryRedirect))
		}
	}

	mux.Handle("GET /api/update/{name}", websocket.Handler(e.handleUpdate))
	mux.HandleFunc("POST /api/files", e.handleUpload)
	mux.HandleFunc("GET /api/app", e.handleAppConf)
	mux.HandleFunc("GET /api/health", e.handleHealth)

	mux.Handle("GET /static/", http.FileServerFS(toolguiweb.GetStaticDir()))

	mux.Handle("GET /assets/", http.StripPrefix("/assets/",
		http.HandlerFunc(e.handleAsset)))

	return mux, nil
}

// StartService start serving the app at addr.
func (e *WebExecutor) StartService(addr string) error {
	mux, err := e.Mux()
	if err != nil {
		return tgutil.Errorf("%w", err)
	}

	err = http.ListenAndServe(addr, mux)
	if err != nil {
		return tgutil.Errorf("%w", err)
	}

	return nil
}
