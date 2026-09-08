package tgexec

import (
	"bufio"
	"bytes"
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

// newTestServer starts a server for an app with a single page.
func newTestServer(t *testing.T) (*httptest.Server, *WebExecutor) {
	t.Helper()

	app := tgframe.NewApp()
	app.AddPage("index", "Index", func(p *tgframe.Params) error { return nil })

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

	return pack.StateID
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
		uploadRequest(t, srv.URL, stateID, "comp", "hello file"))
	if err != nil {
		t.Fatalf("upload: %v", err)
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("StatusCode = %d, want 200", resp.StatusCode)
	}

	state, _ := e.stateMap.Get(stateID)
	file := state.GetFile("comp")
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
		uploadRequest(t, srv.URL, "no_such_state", "comp", "hello"))
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
	req.Header.Set("COMPONENT_ID", "comp")

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
	const componentID = "fileupload_component_File"

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
	e.maxUploadSize = 512

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
	req.Header.Set("COMPONENT_ID", "comp")

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
	req.Header.Set("COMPONENT_ID", "comp")

	resp, err := srv.Client().Do(req)
	if err != nil {
		t.Fatalf("upload: %v", err)
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("StatusCode = %d, want 200", resp.StatusCode)
	}

	state, _ := e.stateMap.Get(stateID)
	bs, err := state.GetFile("comp").Bytes()
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

	e.SetAllowedOrigins([]string{"https://Tools.Example.com/"})

	if _, err := dialUpdateOrigin(t, srv, "https://tools.example.com"); err != nil {
		t.Fatalf("dial: %v", err)
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
}
