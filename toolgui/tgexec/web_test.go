package tgexec

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"testing/fstest"
	"time"

	"github.com/voilelab/toolgui/toolgui/tgcomp/tcinput"
	"github.com/voilelab/toolgui/toolgui/tgframe"
	"github.com/voilelab/toolgui/toolgui/tgutil"

	"golang.org/x/net/websocket"
)

// testComponentID is the fileupload the test page draws. An upload has to
// name a component the page drew, so this is the only id one may carry.
const testComponentID = "fileupload_component_File"

// newTestServer starts a server for an app with a single page.
func newTestServer(t *testing.T) (*httptest.Server, *WebExecutor) {
	t.Helper()

	app := tgframe.NewApp()
	app.AddPage("index", "Index", func(p *tgframe.Params) error {
		tcinput.Fileupload(p.Main, "File", "")
		return nil
	})

	e := NewWebExecutor(app)
	t.Cleanup(e.Destroy)

	mux, err := e.Mux()
	if err != nil {
		t.Fatalf("Mux: %v", err)
	}

	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	return srv, e
}

func dialUpdate(t *testing.T, srv *httptest.Server, pageName string) *websocket.Conn {
	t.Helper()

	url := strings.Replace(srv.URL, "http://", "ws://", 1) + "/api/update/" + pageName
	ws, err := websocket.Dial(url, "", srv.URL)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	t.Cleanup(func() { ws.Close() })

	return ws
}

// A page name the app doesn't have can't start working on a retry, so the
// client is told to stop rather than reconnect in a loop.
func TestUpdateUnknownPageIsFatal(t *testing.T) {
	srv, _ := newTestServer(t)
	ws := dialUpdate(t, srv, "no_such_page")

	var pack tgframe.ResultPack
	if err := websocket.JSON.Receive(ws, &pack); err != nil {
		t.Fatalf("receive: %v", err)
	}

	if pack.Success {
		t.Error("Success = true, want false")
	}

	if pack.Error != "page not found" {
		t.Errorf("Error = %q, want %q", pack.Error, "page not found")
	}

	if !pack.Fatal {
		t.Error("Fatal = false, want true")
	}

	// The server is done with the socket: no session was started for it.
	var bs []byte
	if err := websocket.Message.Receive(ws, &bs); err == nil {
		t.Errorf("socket stayed open, received %q", bs)
	}
}

// A known page has no reason to be fatal: the client should keep the socket.
func TestUpdateKnownPageIsNotFatal(t *testing.T) {
	srv, _ := newTestServer(t)
	ws := dialUpdate(t, srv, "index")

	if err := websocket.JSON.Send(ws, stateIDPack{}); err != nil {
		t.Fatalf("send state id: %v", err)
	}

	var pack stateIDPack
	if err := websocket.JSON.Receive(ws, &pack); err != nil {
		t.Fatalf("receive: %v", err)
	}

	if pack.StateID == "" {
		t.Error("StateID is empty, want a new one")
	}
}

// newUploadState opens an update socket and returns the state ID the server
// hands out for it, which is what an upload has to carry.
func newUploadState(t *testing.T, srv *httptest.Server) string {
	t.Helper()

	ws := dialUpdate(t, srv, "index")
	if err := websocket.JSON.Send(ws, stateIDPack{}); err != nil {
		t.Fatalf("send state id: %v", err)
	}

	var pack stateIDPack
	if err := websocket.JSON.Receive(ws, &pack); err != nil {
		t.Fatalf("receive: %v", err)
	}

	// An empty event is what the client sends once connected. The run it
	// starts is what draws the page's components, and an upload may only name
	// one of those.
	if err := websocket.Message.Send(ws, []byte(`{}`)); err != nil {
		t.Fatalf("send empty event: %v", err)
	}

	waitRun(t, ws)

	return pack.StateID
}

// waitRun reads what the server sends until a run reports its result, so the
// test goes on with the page drawn.
func waitRun(t *testing.T, ws *websocket.Conn) {
	t.Helper()

	if err := ws.SetReadDeadline(time.Now().Add(5 * time.Second)); err != nil {
		t.Fatalf("set deadline: %v", err)
	}
	defer func() { _ = ws.SetReadDeadline(time.Time{}) }()

	for {
		var bs []byte
		if err := websocket.Message.Receive(ws, &bs); err != nil {
			t.Fatalf("receive: %v", err)
		}

		var pack tgframe.ResultPack
		if err := json.Unmarshal(bs, &pack); err != nil {
			continue
		}

		// A notify pack and the ready pack carry neither, so what ends the
		// wait is the result the run finishes with.
		if pack.Success || pack.Error != "" {
			return
		}
	}
}

// uploadRequest builds a multipart upload of content for componentID.
func uploadRequest(t *testing.T, url, stateID, componentID, content string) *http.Request {
	t.Helper()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	part, err := writer.CreateFormFile("file", "a.txt")
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}

	if _, err := part.Write([]byte(content)); err != nil {
		t.Fatalf("write part: %v", err)
	}

	if err := writer.Close(); err != nil {
		t.Fatalf("close writer: %v", err)
	}

	req, err := http.NewRequest(http.MethodPost, url+"/api/files", &body)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("STATE_ID", stateID)
	if componentID != "" {
		req.Header.Set("COMPONENT_ID", componentID)
	}

	return req
}

// TestUploadStoresFileUnderComponent checks an upload lands in the state under
// the component that asked for it, not under the name the browser chose.
func TestUploadStoresFileUnderComponent(t *testing.T) {
	srv, e := newTestServer(t)
	stateID := newUploadState(t, srv)

	resp, err := srv.Client().Do(
		uploadRequest(t, srv.URL, stateID, testComponentID, "hello file"))
	if err != nil {
		t.Fatalf("upload: %v", err)
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("StatusCode = %d, want 200", resp.StatusCode)
	}

	state, _ := e.stateMap.Get(stateID)
	file := state.GetFile(testComponentID)
	if file == nil {
		t.Fatal("expect the upload in the state")
	}

	if file.Name() != "a.txt" {
		t.Errorf("Name = %q, want a.txt", file.Name())
	}

	bs, err := file.Bytes()
	if err != nil {
		t.Fatalf("Bytes: %v", err)
	}

	if string(bs) != "hello file" {
		t.Errorf("Bytes = %q, want hello file", bs)
	}
}

// TestUploadWithUndeclaredComponentID checks an upload naming a component the
// page never drew is refused. Taking one would let a caller keep a file per
// name it invents, and there is nothing that ever reads or releases those.
func TestUploadWithUndeclaredComponentID(t *testing.T) {
	srv, e := newTestServer(t)
	stateID := newUploadState(t, srv)

	resp, err := srv.Client().Do(
		uploadRequest(t, srv.URL, stateID, "no_such_component", "hello"))
	if err != nil {
		t.Fatalf("upload: %v", err)
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("StatusCode = %d, want 403", resp.StatusCode)
	}

	state, _ := e.stateMap.Get(stateID)
	if state.GetFile("no_such_component") != nil {
		t.Error("expect the refused upload to be stored nowhere")
	}
}

// TestUploadWithoutComponentID checks an upload that doesn't say which
// component it belongs to is refused rather than stored somewhere arbitrary.
func TestUploadWithoutComponentID(t *testing.T) {
	srv, _ := newTestServer(t)
	stateID := newUploadState(t, srv)

	resp, err := srv.Client().Do(
		uploadRequest(t, srv.URL, stateID, "", "hello"))
	if err != nil {
		t.Fatalf("upload: %v", err)
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("StatusCode = %d, want 400", resp.StatusCode)
	}
}

func TestUploadWithUnknownStateID(t *testing.T) {
	srv, _ := newTestServer(t)

	resp, err := srv.Client().Do(
		uploadRequest(t, srv.URL, "no_such_state", testComponentID, "hello"))
	if err != nil {
		t.Fatalf("upload: %v", err)
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("StatusCode = %d, want 403", resp.StatusCode)
	}
}

// TestUploadWithoutFilePart checks a multipart body carrying no file is
// refused instead of storing an empty one.
func TestUploadWithoutFilePart(t *testing.T) {
	srv, _ := newTestServer(t)
	stateID := newUploadState(t, srv)

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	if err := writer.WriteField("other", "x"); err != nil {
		t.Fatalf("write field: %v", err)
	}

	if err := writer.Close(); err != nil {
		t.Fatalf("close writer: %v", err)
	}

	req, err := http.NewRequest(http.MethodPost, srv.URL+"/api/files", &body)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("STATE_ID", stateID)
	req.Header.Set("COMPONENT_ID", testComponentID)

	resp, err := srv.Client().Do(req)
	if err != nil {
		t.Fatalf("upload: %v", err)
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("StatusCode = %d, want 400", resp.StatusCode)
	}
}

// TestWriteUploadErrorTooLarge checks an upload over the limit is answered as
// too large, even wrapped the way the store path wraps it.
func TestWriteUploadErrorTooLarge(t *testing.T) {
	err := tgutil.Errorf("%w", &http.MaxBytesError{Limit: MaxUploadSize})

	w := httptest.NewRecorder()
	writeUploadError(w, err)

	if w.Code != http.StatusRequestEntityTooLarge {
		t.Errorf("Code = %d, want 413", w.Code)
	}
}

func TestWriteUploadErrorOther(t *testing.T) {
	w := httptest.NewRecorder()
	writeUploadError(w, tgutil.NewError("disk is full"))

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Code = %d, want 500", w.Code)
	}
}

// TestUploadReachesFileupload runs the whole upload path the browser takes:
// POST the file, then send the pick over the socket, and check the page reads
// the content back.
func TestUploadReachesFileupload(t *testing.T) {
	// The id the browser puts on the input, and the one the component looks
	// its content up under.
	const componentID = testComponentID

	files := make(chan string, 4)

	app := tgframe.NewApp()
	app.AddPage("index", "Index", func(p *tgframe.Params) error {
		fileObj := tcinput.Fileupload(p.Main, "File", "")
		if fileObj == nil {
			files <- ""
			return nil
		}

		bs, err := fileObj.Bytes()
		if err != nil {
			return err
		}

		files <- string(bs)
		return nil
	})

	e := NewWebExecutor(app)
	t.Cleanup(e.Destroy)

	mux, err := e.Mux()
	if err != nil {
		t.Fatalf("Mux: %v", err)
	}

	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	ws := dialUpdate(t, srv, "index")
	if err := websocket.JSON.Send(ws, stateIDPack{}); err != nil {
		t.Fatalf("send state id: %v", err)
	}

	var pack stateIDPack
	if err := websocket.JSON.Receive(ws, &pack); err != nil {
		t.Fatalf("receive: %v", err)
	}

	// An empty event is what the client sends once connected, and what draws
	// the page for the first time.
	if err := websocket.Message.Send(ws, []byte(`{}`)); err != nil {
		t.Fatalf("send empty event: %v", err)
	}

	if got := waitFile(t, files); got != "" {
		t.Fatalf("expect no file on the first run, got %q", got)
	}

	resp, err := srv.Client().Do(
		uploadRequest(t, srv.URL, pack.StateID, componentID, "hello file"))
	if err != nil {
		t.Fatalf("upload: %v", err)
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("StatusCode = %d, want 200", resp.StatusCode)
	}

	event := `{"type":"input","id":"` + componentID +
		`","value":{"name":"a.txt","type":"text/plain","size":10}}`
	if err := websocket.Message.Send(ws, []byte(event)); err != nil {
		t.Fatalf("send event: %v", err)
	}

	if got := waitFile(t, files); got != "hello file" {
		t.Errorf("page read %q, want hello file", got)
	}
}

// waitFile returns what the next page run read, failing the test if no run
// comes.
func waitFile(t *testing.T, files chan string) string {
	t.Helper()

	select {
	case got := <-files:
		return got
	case <-time.After(5 * time.Second):
		t.Fatal("the page did not run")
		return ""
	}
}

// TestUploadOverMaxSize checks the cap covers the request, not just the file
// part: a small file followed by a huge field is still refused.
func TestUploadOverMaxSize(t *testing.T) {
	srv, e := newTestServer(t)
	e.SetMaxUploadSize(512)

	stateID := newUploadState(t, srv)

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	part, err := writer.CreateFormFile("file", "a.txt")
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}

	if _, err := part.Write([]byte("small")); err != nil {
		t.Fatalf("write part: %v", err)
	}

	if err := writer.WriteField("trailing", strings.Repeat("x", 1024)); err != nil {
		t.Fatalf("write field: %v", err)
	}

	if err := writer.Close(); err != nil {
		t.Fatalf("close writer: %v", err)
	}

	req, err := http.NewRequest(http.MethodPost, srv.URL+"/api/files", &body)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("STATE_ID", stateID)
	req.Header.Set("COMPONENT_ID", testComponentID)

	resp, err := srv.Client().Do(req)
	if err != nil {
		t.Fatalf("upload: %v", err)
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusRequestEntityTooLarge {
		t.Errorf("StatusCode = %d, want 413", resp.StatusCode)
	}
}

// TestUploadWithTrailingField checks a body that carries more than the file
// still succeeds once it fits the cap.
func TestUploadWithTrailingField(t *testing.T) {
	srv, e := newTestServer(t)
	stateID := newUploadState(t, srv)

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	part, err := writer.CreateFormFile("file", "a.txt")
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}

	if _, err := part.Write([]byte("hello file")); err != nil {
		t.Fatalf("write part: %v", err)
	}

	if err := writer.WriteField("trailing", "x"); err != nil {
		t.Fatalf("write field: %v", err)
	}

	if err := writer.Close(); err != nil {
		t.Fatalf("close writer: %v", err)
	}

	req, err := http.NewRequest(http.MethodPost, srv.URL+"/api/files", &body)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("STATE_ID", stateID)
	req.Header.Set("COMPONENT_ID", testComponentID)

	resp, err := srv.Client().Do(req)
	if err != nil {
		t.Fatalf("upload: %v", err)
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("StatusCode = %d, want 200", resp.StatusCode)
	}

	state, _ := e.stateMap.Get(stateID)
	bs, err := state.GetFile(testComponentID).Bytes()
	if err != nil {
		t.Fatalf("Bytes: %v", err)
	}

	if string(bs) != "hello file" {
		t.Errorf("Bytes = %q, want hello file", bs)
	}
}

// An app serves its own files -- a manifest icon, a page's image -- from the
// fs it hands the executor.
func TestAssetsAreServed(t *testing.T) {
	srv, e := newTestServer(t)
	e.SetAssets(fstest.MapFS{"icon.png": {Data: []byte("PNG")}})

	resp, err := http.Get(srv.URL + "/assets/icon.png")
	if err != nil {
		t.Fatalf("get asset: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read asset body: %v", err)
	}

	if string(body) != "PNG" {
		t.Errorf("body = %q, want %q", body, "PNG")
	}

	// A file the app doesn't have says so, rather than handing back a page.
	missing, err := http.Get(srv.URL + "/assets/nope.png")
	if err != nil {
		t.Fatalf("get missing asset: %v", err)
	}
	defer missing.Body.Close()

	if missing.StatusCode != http.StatusNotFound {
		t.Errorf("status = %d, want %d", missing.StatusCode, http.StatusNotFound)
	}
}

// An app that gave no fs has no files, rather than a page under /assets/.
func TestAssetsUnsetIsNotFound(t *testing.T) {
	srv, _ := newTestServer(t)

	resp, err := http.Get(srv.URL + "/assets/icon.png")
	if err != nil {
		t.Fatalf("get asset: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusNotFound)
	}
}

// The setters promise the app can call them whenever, so a call that lands
// while requests are in flight must not race the handlers reading them.
// Meaningful under -race.
func TestSetManifestAndAssetsAreConcurrencySafe(t *testing.T) {
	srv, e := newTestServer(t)

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()

		for i := range 50 {
			if i%2 == 0 {
				e.SetManifest(&Manifest{Name: "My Tool"})
				e.SetAssets(fstest.MapFS{"icon.png": {Data: []byte("PNG")}})
			} else {
				e.SetManifest(nil)
				e.SetAssets(nil)
			}
		}
	}()

	for _, path := range []string{"/manifest.json", "/assets/icon.png"} {
		wg.Add(1)

		go func() {
			defer wg.Done()

			for range 50 {
				resp, err := http.Get(srv.URL + path)
				if err != nil {
					t.Errorf("get %s: %v", path, err)
					return
				}

				if _, err := io.Copy(io.Discard, resp.Body); err != nil {
					t.Errorf("read %s: %v", path, err)
				}

				resp.Body.Close()
			}
		}()
	}

	wg.Wait()
}

// A plugin's files are served under the prefix its urls are built from, next
// to the app's own routes.
func TestMuxServesPluginAssets(t *testing.T) {
	app := tgframe.NewApp()
	app.AddPage("index", "Index", func(p *tgframe.Params) error { return nil })

	err := app.AddPluginAssets("gauge", fstest.MapFS{
		"gauge.js": &fstest.MapFile{Data: []byte("window.toolgui.update(1)")},
	})
	if err != nil {
		t.Fatalf("AddPluginAssets: %v", err)
	}

	e := NewWebExecutor(app)
	t.Cleanup(e.Destroy)

	mux, err := e.Mux()
	if err != nil {
		t.Fatalf("Mux: %v", err)
	}

	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	resp, err := http.Get(srv.URL + tgframe.PluginAssetURL("gauge", "gauge.js"))
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read: %v", err)
	}

	if string(body) != "window.toolgui.update(1)" {
		t.Errorf("body = %q", body)
	}
}

// dialUpdateOrigin dials the update socket with an Origin of its own, so a
// test can play a page that the app doesn't belong to.
func dialUpdateOrigin(t *testing.T, srv *httptest.Server, origin string) (*websocket.Conn, error) {
	t.Helper()

	url := strings.Replace(srv.URL, "http://", "ws://", 1) + "/api/update/index"

	ws, err := websocket.Dial(url, "", origin)
	if ws != nil {
		t.Cleanup(func() { ws.Close() })
	}

	return ws, err
}

// The same-origin policy doesn't cover websockets, so a page on another site
// can reach the socket. It must not get a session out of it.
func TestUpdateRejectsCrossSiteOrigin(t *testing.T) {
	srv, e := newTestServer(t)

	if _, err := dialUpdateOrigin(t, srv, "http://evil.example"); err == nil {
		t.Fatal("dial succeeded, want a refused handshake")
	}

	if size := e.stateMap.Size(); size != 0 {
		t.Errorf("states = %d, want 0", size)
	}
}

// The app's own pages carry an Origin matching the host they were served
// from, whichever name and port that is.
func TestUpdateAcceptsSameHostOrigin(t *testing.T) {
	srv, _ := newTestServer(t)

	// httptest serves on 127.0.0.1, and localhost reaches the same port under
	// a name of its own.
	hosts := []string{srv.URL, strings.Replace(srv.URL, "127.0.0.1", "localhost", 1)}

	for _, host := range hosts {
		url := strings.Replace(host, "http://", "ws://", 1) + "/api/update/index"

		ws, err := websocket.Dial(url, "", host)
		if err != nil {
			t.Fatalf("dial %s: %v", host, err)
		}

		ws.Close()
	}
}

// A reverse proxy serves the app under an origin of its own, which the app
// has to name for the browser to get through.
func TestUpdateAcceptsAllowedOrigin(t *testing.T) {
	srv, e := newTestServer(t)

	// Case and trailing slashes are noise an entry may carry; the origin a
	// browser sends has neither.
	e.SetAllowedOrigins([]string{
		"https://Tools.Example.com/",
		"https://proxy.example.com//",
	})

	for _, origin := range []string{"https://tools.example.com", "https://proxy.example.com"} {
		if _, err := dialUpdateOrigin(t, srv, origin); err != nil {
			t.Fatalf("dial %s: %v", origin, err)
		}
	}

	if _, err := dialUpdateOrigin(t, srv, "https://other.example.com"); err == nil {
		t.Error("dial succeeded, want a refused handshake")
	}
}

// A client that sends no Origin at all is refused, as it was before the host
// check went in. x/net/websocket's own client won't dial without one, so the
// handshake goes out by hand.
func TestUpdateRejectsMissingOrigin(t *testing.T) {
	srv, _ := newTestServer(t)

	host := strings.TrimPrefix(srv.URL, "http://")

	conn, err := net.Dial("tcp", host)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer conn.Close()

	fmt.Fprintf(conn, "GET /api/update/index HTTP/1.1\r\n"+
		"Host: %s\r\n"+
		"Upgrade: websocket\r\n"+
		"Connection: Upgrade\r\n"+
		"Sec-WebSocket-Key: dGhlIHNhbXBsZSBub25jZQ==\r\n"+
		"Sec-WebSocket-Version: 13\r\n\r\n", host)

	resp, err := http.ReadResponse(bufio.NewReader(conn), nil)
	if err != nil {
		t.Fatalf("read response: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusForbidden)
	}

	// A refused handshake says only that it was refused: x/net/websocket
	// keeps the error out of the response, so nothing of the server leaks.
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}

	if len(body) != 0 {
		t.Errorf("body = %q, want empty", body)
	}
}

// dialUpdateRaw opens an update socket over a TCP connection the test holds,
// so it can drop it the way a network does rather than closing it politely.
func dialUpdateRaw(t *testing.T, srv *httptest.Server) (*websocket.Conn, *net.TCPConn) {
	t.Helper()

	url := strings.Replace(srv.URL, "http://", "ws://", 1) + "/api/update/index"
	config, err := websocket.NewConfig(url, srv.URL)
	if err != nil {
		t.Fatalf("config: %v", err)
	}

	conn, err := net.Dial("tcp", srv.Listener.Addr().String())
	if err != nil {
		t.Fatalf("dial: %v", err)
	}

	ws, err := websocket.NewClient(config, conn)
	if err != nil {
		conn.Close()
		t.Fatalf("websocket client: %v", err)
	}

	return ws, conn.(*net.TCPConn)
}

// TestUpdateResetReleasesState checks a connection that breaks without a
// close -- a reset, what a dropped network looks like -- gives its state back.
// The read loop used to carry on past anything that wasn't io.EOF, which left
// the state alive for good and spun on an error that never goes away.
func TestUpdateResetReleasesState(t *testing.T) {
	srv, e := newTestServer(t)

	ws, conn := dialUpdateRaw(t, srv)

	if err := websocket.JSON.Send(ws, stateIDPack{}); err != nil {
		t.Fatalf("send state id: %v", err)
	}

	var pack stateIDPack
	if err := websocket.JSON.Receive(ws, &pack); err != nil {
		t.Fatalf("receive: %v", err)
	}

	state, alive := e.stateMap.Get(pack.StateID)
	if state == nil || !alive {
		t.Fatal("expect a live state for the connection")
	}

	state.Set("key", "value")

	// Linger 0 makes the close a reset, so the server's next read fails with
	// something other than io.EOF.
	if err := conn.SetLinger(0); err != nil {
		t.Fatalf("set linger: %v", err)
	}
	conn.Close()

	waitNotAlive(t, e, pack.StateID)

	// The state is kept for the reconnect, with what the page put in it.
	back, alive := e.stateMap.Get(pack.StateID)
	if back != state {
		t.Fatal("expect the state to be kept for a reconnect")
	}

	if alive {
		t.Error("expect the state to be back to not alive")
	}

	if v, _ := back.Get[string]("key"); v != "value" {
		t.Errorf("key = %q, want value", v)
	}

	// A reconnect naming it takes it back, rather than being turned away for
	// a connection that is long gone.
	ws2 := dialUpdate(t, srv, "index")
	if err := websocket.JSON.Send(ws2, stateIDPack{StateID: pack.StateID}); err != nil {
		t.Fatalf("send state id: %v", err)
	}

	if err := websocket.Message.Send(ws2, []byte(`{}`)); err != nil {
		t.Fatalf("send empty event: %v", err)
	}

	waitRun(t, ws2)

	if _, alive := e.stateMap.Get(pack.StateID); !alive {
		t.Error("expect the reconnect to take the state back")
	}
}

// waitNotAlive waits for the server to hand the state back.
func waitNotAlive(t *testing.T, e *WebExecutor, stateID string) {
	t.Helper()

	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if _, alive := e.stateMap.Get(stateID); !alive {
			return
		}

		time.Sleep(5 * time.Millisecond)
	}

	t.Fatal("the state stayed alive after the connection broke")
}

// TestUpdateOverMaxStateCount checks a connection asking for a state the
// service has no room for is told so, rather than being handed one anyway.
func TestUpdateOverMaxStateCount(t *testing.T) {
	srv, e := newTestServer(t)
	e.SetMaxStateCount(1)

	newUploadState(t, srv)

	ws := dialUpdate(t, srv, "index")
	if err := websocket.JSON.Send(ws, stateIDPack{}); err != nil {
		t.Fatalf("send state id: %v", err)
	}

	var pack tgframe.ResultPack
	if err := websocket.JSON.Receive(ws, &pack); err != nil {
		t.Fatalf("receive: %v", err)
	}

	if pack.Success {
		t.Error("Success = true, want false")
	}

	if pack.Error == "" {
		t.Error("expect the refusal to say why")
	}

	// The client may come back once a connection frees a state, so this is
	// not the end of the road for it.
	if pack.Fatal {
		t.Error("Fatal = true, want false")
	}
}
