package tgexec

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/voilelab/toolgui/toolgui/tgcomp/tcinput"
	"github.com/voilelab/toolgui/toolgui/tgframe"
	"github.com/voilelab/toolgui/toolgui/tgjson"

	"golang.org/x/net/websocket"
)

// downloadBodySize is what the test page offers: past the chunk a copy moves
// at a time, so a body that stopped at the first chunk would not pass here.
const downloadBodySize = 1 << 20

// downloadBody is the file the pages below offer. A pattern rather than noise,
// so a byte anywhere in it is known from its offset alone.
func downloadBody() []byte {
	bs := make([]byte, downloadBodySize)
	for i := range bs {
		bs[i] = byte(i % 251)
	}

	return bs
}

// newDownloadServer starts a server for a page that offers what body yields,
// once per run.
func newDownloadServer(t *testing.T,
	body func() []byte) (*httptest.Server, *WebExecutor) {
	t.Helper()

	app := tgframe.NewApp()
	app.AddPage("index", "Index", func(p *tgframe.Params) error {
		tcinput.DownloadFile(p.Main, "Download", body(),
			&tcinput.DownloadFileConf{Filename: "pattern.bin", MIME: "text/csv"})
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

// openDownloadPage opens a connection of its own and returns it with the state
// id the server gave it.
func openDownloadPage(t *testing.T,
	srv *httptest.Server) (*websocket.Conn, string) {
	t.Helper()

	ws := dialUpdate(t, srv, "index")
	if err := jsonCodec.Send(ws, stateIDPack{}); err != nil {
		t.Fatalf("send state id: %v", err)
	}

	var pack stateIDPack
	if err := jsonCodec.Receive(ws, &pack); err != nil {
		t.Fatalf("receive: %v", err)
	}

	return ws, pack.StateID
}

// drawDownloadPage runs the page and returns the token the download
// component's pack carries, which is the only place a client ever reads one
// from.
func drawDownloadPage(t *testing.T, ws *websocket.Conn) string {
	t.Helper()

	if err := websocket.Message.Send(ws, []byte(`{}`)); err != nil {
		t.Fatalf("send empty event: %v", err)
	}

	if err := ws.SetReadDeadline(time.Now().Add(10 * time.Second)); err != nil {
		t.Fatalf("set deadline: %v", err)
	}
	defer func() { _ = ws.SetReadDeadline(time.Time{}) }()

	token := ""

	for {
		var bs []byte
		if err := websocket.Message.Receive(ws, &bs); err != nil {
			t.Fatalf("receive: %v", err)
		}

		var msg struct {
			Success   bool   `json:"success"`
			Error     string `json:"error"`
			Component struct {
				Name  string `json:"name"`
				Token string `json:"token"`
			} `json:"component"`
		}

		if err := tgjson.Unmarshal(bs, &msg); err != nil {
			continue
		}

		if msg.Component.Name == "download_file_component" {
			token = msg.Component.Token
		}

		if msg.Success || msg.Error != "" {
			break
		}
	}

	if token == "" {
		t.Fatal("the run drew no download component")
	}

	return token
}

// newDownloadPage is a server, a connection that has drawn the page and the
// token that run offered.
func newDownloadPage(t *testing.T) (*httptest.Server, string, string) {
	t.Helper()

	srv, _ := newDownloadServer(t, downloadBody)
	ws, stateID := openDownloadPage(t, srv)

	return srv, stateID, drawDownloadPage(t, ws)
}

// downloadRequest builds a fetch of token against stateID, the way the client
// makes one: both in headers, neither in the URL.
func downloadRequest(t *testing.T, srv *httptest.Server, stateID,
	token string) *http.Request {
	t.Helper()

	req, err := http.NewRequest(http.MethodGet, srv.URL+"/api/files", nil)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}

	if stateID != "" {
		req.Header.Set("STATE_ID", stateID)
	}

	if token != "" {
		req.Header.Set("DOWNLOAD_TOKEN", token)
	}

	return req
}

// TestDownloadServesWhatThePageOffered is the endpoint's job: the bytes the
// run offered come back whole, under the name and the type it offered them as.
func TestDownloadServesWhatThePageOffered(t *testing.T) {
	srv, stateID, token := newDownloadPage(t)

	resp, err := srv.Client().Do(downloadRequest(t, srv, stateID, token))
	if err != nil {
		t.Fatalf("download: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("StatusCode = %d, want 200", resp.StatusCode)
	}

	if got := resp.Header.Get("Content-Type"); got != "text/csv" {
		t.Errorf("Content-Type = %q, want text/csv", got)
	}

	if got := resp.Header.Get("Content-Disposition"); !strings.Contains(got,
		"pattern.bin") {
		t.Errorf("Content-Disposition = %q, want the filename in it", got)
	}

	if got := resp.Header.Get("X-Content-Type-Options"); got != "nosniff" {
		t.Errorf("X-Content-Type-Options = %q, want nosniff", got)
	}

	if got := resp.Header.Get("Cache-Control"); got != "no-store" {
		t.Errorf("Cache-Control = %q, want no-store", got)
	}

	var got bytes.Buffer
	if _, err := got.ReadFrom(resp.Body); err != nil {
		t.Fatalf("read body: %v", err)
	}

	if !bytes.Equal(got.Bytes(), downloadBody()) {
		t.Errorf("served %d bytes, want the %d the page offered, identical",
			got.Len(), downloadBodySize)
	}
}

// TestDownloadWithoutToken checks a state id alone fetches nothing.
func TestDownloadWithoutToken(t *testing.T) {
	srv, stateID, _ := newDownloadPage(t)

	resp, err := srv.Client().Do(downloadRequest(t, srv, stateID, ""))
	if err != nil {
		t.Fatalf("download: %v", err)
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("StatusCode = %d, want 400", resp.StatusCode)
	}
}

// TestDownloadWithUnknownToken checks a token nothing handed out fetches
// nothing, which is the only answer a caller guessing at tokens gets.
func TestDownloadWithUnknownToken(t *testing.T) {
	srv, stateID, _ := newDownloadPage(t)

	resp, err := srv.Client().Do(
		downloadRequest(t, srv, stateID, "NOTATOKENNOTATOKENNOTATOKEN"))
	if err != nil {
		t.Fatalf("download: %v", err)
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("StatusCode = %d, want 404", resp.StatusCode)
	}
}

// TestDownloadWithoutStateID checks the token is a bearer of nothing on its
// own: with no connection to look it up in, there is nothing to serve.
func TestDownloadWithoutStateID(t *testing.T) {
	srv, _, token := newDownloadPage(t)

	resp, err := srv.Client().Do(downloadRequest(t, srv, "", token))
	if err != nil {
		t.Fatalf("download: %v", err)
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("StatusCode = %d, want 403", resp.StatusCode)
	}
}

// TestDownloadAcrossStates is the part worth pinning: a token is looked up in
// the state that offered it and nowhere else, so one page's output cannot be
// read through another page's connection.
func TestDownloadAcrossStates(t *testing.T) {
	srv, _ := newDownloadServer(t, downloadBody)

	mine, _ := openDownloadPage(t, srv)
	token := drawDownloadPage(t, mine)

	// Another connection to the same server, with a state of its own.
	_, theirs := openDownloadPage(t, srv)

	resp, err := srv.Client().Do(downloadRequest(t, srv, theirs, token))
	if err != nil {
		t.Fatalf("download: %v", err)
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("StatusCode = %d, want 404", resp.StatusCode)
	}
}

// TestDownloadOfARetiredToken checks a run that offers a different file
// retires the token before it: what a token names is one run's output, not
// whatever the component holds later.
func TestDownloadOfARetiredToken(t *testing.T) {
	runs := 0
	srv, _ := newDownloadServer(t, func() []byte {
		runs++
		return []byte(strings.Repeat("run", runs))
	})

	ws, stateID := openDownloadPage(t, srv)
	token := drawDownloadPage(t, ws)

	// A rerun with something else to offer.
	if next := drawDownloadPage(t, ws); next == token {
		t.Fatal("expect a new token for the new file")
	}

	resp, err := srv.Client().Do(downloadRequest(t, srv, stateID, token))
	if err != nil {
		t.Fatalf("download: %v", err)
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("StatusCode = %d, want 404 for the token of the run before",
			resp.StatusCode)
	}
}
