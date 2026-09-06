package tgexec

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/voilelab/toolgui/toolgui/tgframe"

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
