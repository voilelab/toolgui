package tgexec

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
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
		fileObj := tcinput.Fileupload(p.State, p.Main, "File", "")
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
