package tgexec

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"io/fs"
	"log/slog"
	"mime"
	"strconv"
	"strings"
	"sync"
	"time"

	"net/http"
	"net/url"

	toolguiweb "github.com/voilelab/toolgui/toolgui-web"
	"github.com/voilelab/toolgui/toolgui/tgframe"
	"github.com/voilelab/toolgui/toolgui/tgjson"
	"github.com/voilelab/toolgui/toolgui/tgutil"

	"golang.org/x/net/websocket"
)

// MaxUploadSize limit the size of a file upload request.
const MaxUploadSize int64 = 1024 * 1024 * 1024

// MaxMessageSize caps one update message, which is held in memory and
// parsed from an unauthenticated connection.
const MaxMessageSize = 1024 * 1024

// DefaultMaxStateCount is how many states an executor keeps by default. Each
// unauthenticated connection takes one, so it must be capped.
const DefaultMaxStateCount = 1024

// stateIDTimeout is how long a connection has after the handshake to send
// its state id.
const stateIDTimeout = 30 * time.Second

// readHeaderTimeout bounds reading request headers; idleTimeout bounds an
// idle kept-alive connection.
const (
	readHeaderTimeout = 30 * time.Second
	idleTimeout       = 120 * time.Second
)

// ErrUpdateInterrupt is raise at panic when current state is going to interrupt
//
// Deprecated: use [tgframe.ErrUpdateInterrupt].
var ErrUpdateInterrupt = tgframe.ErrUpdateInterrupt

// WebExecutor is a web ui executor for ToolGUI.
type WebExecutor struct {
	rootAssets map[string][]byte

	// index is the page served for every app page.
	index string

	stateMap tgutil.UUIDMap[tgframe.State]

	// maxUploadSize is the size cap of one upload request, guarded by confMu.
	maxUploadSize int64

	// maxMessageSize is the size cap of one update message, guarded by confMu.
	maxMessageSize int

	app *tgframe.App

	// confMu guards the settings below, which may change while serving.
	confMu sync.RWMutex

	// manifest is nil until the app sets one, and nil serves the default.
	manifest *Manifest

	// assets is nil until the app sets one.
	assets fs.FS

	// headHTML is inserted at the start of the index page head.
	headHTML string

	// allowedOrigins are extra normalized origins the update socket accepts.
	allowedOrigins []string
}

type stateIDPack struct {
	StateID string `json:"state_id"`
}

// NewWebExecutor return a WebExecutor.
func NewWebExecutor(app *tgframe.App) *WebExecutor {
	stateMap := tgutil.NewUUIDMap(
		tgframe.NewState, func(t *tgframe.State) { t.Destroy() },
		5*time.Minute)
	stateMap.SetMaxSize(DefaultMaxStateCount)

	return &WebExecutor{
		rootAssets: toolguiweb.GetRootAssets(),
		index:      toolguiweb.IndexBody,

		stateMap: stateMap,

		maxUploadSize:  MaxUploadSize,
		maxMessageSize: MaxMessageSize,

		app: app,
	}
}

// SetMaxStateCount limits how many states (open pages) the executor keeps;
// connections past it are turned away. 0 or less is no limit.
func (e *WebExecutor) SetMaxStateCount(n int) {
	e.stateMap.SetMaxSize(n)
}

// SetMaxUploadSize limits one whole upload request, [MaxUploadSize] by
// default. Larger uploads get 413.
func (e *WebExecutor) SetMaxUploadSize(n int64) {
	e.confMu.Lock()
	defer e.confMu.Unlock()

	e.maxUploadSize = n
}

// SetMaxMessageSize limits one update message, [MaxMessageSize] by default.
// Larger messages are refused without being read; the connection stays open.
func (e *WebExecutor) SetMaxMessageSize(n int) {
	e.confMu.Lock()
	defer e.confMu.Unlock()

	e.maxMessageSize = n
}

// defaultManifest returns the manifest served when the app sets none, named
// after the app title.
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

// SetAssets serves fsys under /assets/, e.g. a manifest icon or images. A
// nil fsys serves none.
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

// SetHeadHTML inserts html into the index page head, for tags that must be
// static: Open Graph, CSP, fonts, analytics.
//
//	e.SetHeadHTML(`<meta property="og:title" content="My Tool" />`)
func (e *WebExecutor) SetHeadHTML(html string) {
	e.confMu.Lock()
	defer e.confMu.Unlock()

	e.headHTML = html
}

// indexBody return the index page with the head html in it.
func (e *WebExecutor) indexBody() []byte {
	e.confMu.RLock()
	head := e.headHTML
	e.confMu.RUnlock()

	if head == "" {
		return []byte(e.index)
	}

	html, _ := tgutil.InsertHead(e.index, head)
	return []byte(html)
}

// SetAllowedOrigins lets pages from these origins (with scheme) open the
// update websocket, besides the app's own:
//
//	e.SetAllowedOrigins([]string{"https://tools.example.com"})
//
// By default only an Origin matching the request Host is accepted. Set the
// public origin here when a reverse proxy rewrites Host.
func (e *WebExecutor) SetAllowedOrigins(origins []string) {
	normalized := make([]string, 0, len(origins))
	for _, origin := range origins {
		normalized = append(normalized, normalizeOrigin(origin))
	}

	e.confMu.Lock()
	defer e.confMu.Unlock()

	e.allowedOrigins = normalized
}

// normalizeOrigin lowercases origin and trims trailing slashes.
func normalizeOrigin(origin string) string {
	return strings.ToLower(strings.TrimRight(origin, "/"))
}

// checkUpdateOrigin rejects update handshakes from foreign origins, since
// the same-origin policy doesn't cover websockets. A rejection gets 403
// before any state is made.
func (e *WebExecutor) checkUpdateOrigin(config *websocket.Config, req *http.Request) error {
	origin, err := websocket.Origin(config, req)
	if err != nil {
		slog.Error("websocket origin", "error", err)
		return err
	}

	if origin == nil {
		// Same as the default handshake.
		return errors.New("websocket: no origin")
	}

	config.Origin = origin

	// The requested host is the app's own origin.
	if strings.EqualFold(origin.Host, req.Host) {
		return nil
	}

	e.confMu.RLock()
	allowed := e.allowedOrigins
	e.confMu.RUnlock()

	want := normalizeOrigin(origin.Scheme + "://" + origin.Host)
	for _, o := range allowed {
		if o == want {
			return nil
		}
	}

	slog.Error("websocket origin not allowed",
		"origin", origin.String(), "host", req.Host)

	return errors.New("websocket: origin not allowed")
}

// stateTag return a stable, non-reversible tag of a state id for logs.
func stateTag(id string) string {
	sum := sha256.Sum256([]byte(id))
	return hex.EncodeToString(sum[:4])
}

// Destory release all resource.
func (e *WebExecutor) Destroy() {
	e.stateMap.Destroy()
}

func (e *WebExecutor) handleUpdate(ws *websocket.Conn) {
	// Oversized frames are refused by header, before reaching memory.
	e.confMu.RLock()
	ws.MaxPayloadBytes = e.maxMessageSize
	e.confMu.RUnlock()

	pageName := ws.Request().PathValue("name")
	if !e.app.HasPage(pageName) {
		jsonCodec.Send(ws, &tgframe.ResultPack{
			Error:   "page not found",
			Success: false,
			Fatal:   true,
		})
		slog.Error("page not found", "page", pageName)
		return
	}

	// The page query is on the socket url. Its values stay out of the log.
	query, err := tgframe.ParseQuery(ws.Request().URL.RawQuery)
	if err != nil {
		jsonCodec.Send(ws, &tgframe.ResultPack{
			Error:   err.Error(),
			Success: false,
			Fatal:   true,
		})
		slog.Error("page query", "page", pageName, "error", err)
		return
	}

	// Deadline the first message so a silent connection can't hold
	// resources; lifted after, since the socket idles between events.
	if err := ws.SetReadDeadline(time.Now().Add(stateIDTimeout)); err != nil {
		slog.Error("set state id deadline", "error", err)
	}

	var pack stateIDPack
	err = jsonCodec.Receive(ws, &pack)

	if derr := ws.SetReadDeadline(time.Time{}); derr != nil {
		slog.Error("clear state id deadline", "error", derr)
	}

	if err != nil {
		jsonCodec.Send(ws, tgframe.ReportError("state id", err))
		return
	}

	stateID := pack.StateID

	// Acquire is atomic, so two connections can't share one idle state.
	state, err := e.stateMap.Acquire(stateID)
	switch {
	case err == nil:
		// Took over an idle state.

	case errors.Is(err, tgutil.ErrUUIDAlive):
		jsonCodec.Send(ws, &tgframe.ResultPack{
			Error:   "state id already alive",
			Success: false,
		})
		// Log a tag, not the id: the id alone takes over the session.
		slog.Error("state id already alive", "state", stateTag(stateID))
		return

	default:
		newStateID, nerr := e.stateMap.New()
		if nerr != nil {
			// At the state cap; retryable once another connection drops.
			jsonCodec.Send(ws, &tgframe.ResultPack{
				Error:   "too many sessions, try again later",
				Success: false,
			})
			slog.Error("new state", "error", nerr)
			return
		}

		stateID = newStateID
		state, _ = e.stateMap.Get(stateID)
		jsonCodec.Send(ws, stateIDPack{
			StateID: stateID,
		})
	}

	session, err := tgframe.NewSession(e.app, pageName, query, state,
		func(pack any) error { return jsonCodec.Send(ws, pack) })
	if err != nil {
		// Checked above, so this is fatal. Release the state.
		e.stateMap.SetAlive(stateID, false)

		msg := "page not found"
		if errors.Is(err, tgframe.ErrQueryTooLarge) {
			msg = tgframe.ErrQueryTooLarge.Error()
		}

		jsonCodec.Send(ws, &tgframe.ResultPack{
			Error:   msg,
			Success: false,
			Fatal:   true,
		})
		slog.Error("new session", "error", err)
		return
	}

	// On exit, stop the session and release the state for reconnect or
	// cleanup.
	defer func() {
		session.Close()
		e.stateMap.SetAlive(stateID, false)
	}()

	for {
		var bs []byte
		err := websocket.Message.Receive(ws, &bs)
		if err != nil {
			if !recoverableReceiveErr(err) {
				// Connection closed or broken.
				if !errors.Is(err, io.EOF) {
					slog.Error("receive", "error", err)
				}

				break
			}

			// Only an oversized frame is recoverable.
			jsonCodec.Send(ws, &tgframe.ResultPack{
				Error:   "message too large",
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

// recoverableReceiveErr reports whether the read loop can continue after
// err. Only a frame over [WebExecutor.SetMaxMessageSize] is: the next Receive
// drains it. Other errors repeat forever.
func recoverableReceiveErr(err error) bool {
	return errors.Is(err, websocket.ErrFrameTooLarge)
}

func (e *WebExecutor) handleUpload(w http.ResponseWriter, req *http.Request) {
	stateID := req.Header.Get("STATE_ID")
	state, alive := e.stateMap.Get(stateID)
	if state == nil || !alive {
		http.Error(w, "State ID is invalid or not alive", http.StatusForbidden)
		return
	}

	// Stored per component. The id is percent-encoded for non-ASCII labels.
	componentID, err := url.PathUnescape(req.Header.Get("COMPONENT_ID"))
	if err != nil {
		http.Error(w, "Component ID is not percent-encoded", http.StatusBadRequest)
		return
	}

	if componentID == "" {
		http.Error(w, "Component ID is missing", http.StatusBadRequest)
		return
	}

	// Reject ids the page never drew, or a caller could fill the disk.
	if !state.HasFileKey(componentID) {
		http.Error(w, "Component ID is not on the page", http.StatusForbidden)
		return
	}

	e.confMu.RLock()
	maxUploadSize := e.maxUploadSize
	e.confMu.RUnlock()

	req.Body = http.MaxBytesReader(w, req.Body, maxUploadSize)

	// Stream parts; ParseMultipartForm would buffer the whole upload.
	reader, err := req.MultipartReader()
	if err != nil {
		http.Error(w, "Not a multipart upload", http.StatusBadRequest)
		slog.Error("Multipart reader", "error", err)
		return
	}

	stored := false

	for {
		part, err := reader.NextPart()
		if err == io.EOF {
			break
		}

		if err != nil {
			writeUploadError(w, err)
			return
		}

		if part.FormName() != "file" || stored {
			// Drain other parts so the size cap covers the whole body.
			_, err = io.Copy(io.Discard, part)
			part.Close()

			if err != nil {
				writeUploadError(w, err)
				return
			}

			continue
		}

		// Streamed straight to storage.
		_, err = state.WriteFile(componentID, part.FileName(), part)
		part.Close()

		if err != nil {
			writeUploadError(w, err)
			return
		}

		stored = true
	}

	if !stored {
		http.Error(w, "Upload has no file part", http.StatusBadRequest)
	}
}

// writeUploadError answers a failed upload with 413 or 500.
func writeUploadError(w http.ResponseWriter, err error) {
	var maxBytesErr *http.MaxBytesError
	if errors.As(err, &maxBytesErr) {
		http.Error(w, "Upload is too large", http.StatusRequestEntityTooLarge)
		slog.Error("Upload too large", "limit", maxBytesErr.Limit)
		return
	}

	http.Error(w, "Store upload failed", http.StatusInternalServerError)
	slog.Error("Store upload", "error", err)
}

// handleDownload serves a file offered through
// [github.com/voilelab/toolgui/toolgui/tgcomp/tcinput.DownloadFile].
//
// It needs both the state id and a token from that state's own downloads.
// Both are headers, not URL parts, to keep them out of logs and Referer.
func (e *WebExecutor) handleDownload(w http.ResponseWriter, req *http.Request) {
	stateID := req.Header.Get("STATE_ID")
	state, alive := e.stateMap.Get(stateID)
	if state == nil || !alive {
		http.Error(w, "State ID is invalid or not alive", http.StatusForbidden)
		return
	}

	token := req.Header.Get("DOWNLOAD_TOKEN")
	if token == "" {
		http.Error(w, "Download token is missing", http.StatusBadRequest)
		return
	}

	download := state.GetDownload(token)
	if download == nil {
		// Foreign, replaced and unknown tokens look the same.
		http.Error(w, "No such download", http.StatusNotFound)
		return
	}

	fp, err := download.Open()
	if err != nil {
		http.Error(w, "Read download failed", http.StatusInternalServerError)
		slog.Error("Open download", "error", err)
		return
	}
	defer fp.Close()

	w.Header().Set("Content-Type", download.MIME())
	w.Header().Set("Content-Length", strconv.FormatInt(download.Size(), 10))

	// For direct fetches (the client uses the pack's name): attachment name,
	// no sniffing, no caching.
	disposition := mime.FormatMediaType("attachment", map[string]string{
		"filename": download.Name(),
	})
	if disposition == "" {
		// Unencodable name; still never serve inline.
		disposition = "attachment"
	}

	w.Header().Set("Content-Disposition", disposition)
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Cache-Control", "no-store")

	// Stream; the file may be too big to hold.
	if _, err := io.Copy(w, fp); err != nil {
		// Headers are already sent; just log.
		slog.Error("Send download", "error", err)
	}
}

func (e *WebExecutor) handlePage(resp http.ResponseWriter, req *http.Request) {
	pageName := req.PathValue("name")
	body, isRootAssets := e.rootAssets[pageName]
	if isRootAssets && pageName != "index.html" {
		resp.Write(body)
		return
	}

	resp.Write(e.indexBody())
}

func (e *WebExecutor) handleAssets(resp http.ResponseWriter, req *http.Request) {
	pageName := req.PathValue("name")
	if pageName == "index.html" {
		resp.Write(e.indexBody())
		return
	}

	body, isRootAssets := e.rootAssets[pageName]
	if !isRootAssets {
		resp.WriteHeader(http.StatusNotFound)
		return
	}

	resp.Write(body)
}

// handleAsset serves [WebExecutor.SetAssets] files, read per request so
// they can be set any time.
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
	resp.Write(e.indexBody())
}

func (e *WebExecutor) handleManifest(resp http.ResponseWriter, req *http.Request) {
	e.confMu.RLock()
	manifest := e.manifest
	e.confMu.RUnlock()

	if manifest == nil {
		manifest = e.defaultManifest()
	}

	bs, err := tgjson.Marshal(manifest)
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
	bs, err := tgjson.Marshal(e.app.AppConf())
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

	// More specific than the page patterns below, so it wins.
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

	mux.Handle("GET /api/update/{name}", &websocket.Server{
		Handler:   e.handleUpdate,
		Handshake: e.checkUpdateOrigin,
	})
	mux.HandleFunc("POST /api/files", e.handleUpload)
	mux.HandleFunc("GET /api/files", e.handleDownload)
	mux.HandleFunc("GET /api/app", e.handleAppConf)
	mux.HandleFunc("GET /api/health", e.handleHealth)

	mux.Handle("GET /static/", http.FileServerFS(toolguiweb.GetStaticDir()))

	mux.Handle("GET /assets/", http.StripPrefix("/assets/",
		http.HandlerFunc(e.handleAsset)))

	mux.Handle("GET "+tgframe.PluginAssetPrefix, tgframe.PluginAssetHandler(e.app))

	return mux, nil
}

// StartService start serving the app at addr.
func (e *WebExecutor) StartService(addr string) error {
	mux, err := e.Mux()
	if err != nil {
		return tgutil.Errorf("%w", err)
	}

	// Deadline slow headers and idle keep-alives. ReadTimeout/WriteTimeout
	// stay unset: they would cut the long-lived update websocket, which
	// handleUpdate bounds itself.
	srv := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: readHeaderTimeout,
		IdleTimeout:       idleTimeout,
	}

	err = srv.ListenAndServe()
	if err != nil {
		return tgutil.Errorf("%w", err)
	}

	return nil
}
