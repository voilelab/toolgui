package tgexec

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"testing/fstest"

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
