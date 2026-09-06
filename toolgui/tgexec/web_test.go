package tgexec

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/voilelab/toolgui/toolgui/tgframe"

	"golang.org/x/net/websocket"
)

// A client asking for a page the app doesn't have has to be told to stop, or
// it reconnects forever.
func TestUpdateRejectsUnknownPageAsFatal(t *testing.T) {
	app := tgframe.NewApp()
	app.AddPage("index", "Index", func(p *tgframe.Params) error { return nil })

	e := NewWebExecutor(app)
	defer e.Destroy()

	mux, err := e.Mux()
	if err != nil {
		t.Fatal(err)
	}

	server := httptest.NewServer(mux)
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	conn, err := websocket.Dial(wsURL+"/api/update/main", "", server.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	var pack tgframe.ResultPack
	if err := websocket.JSON.Receive(conn, &pack); err != nil {
		t.Fatal(err)
	}

	if pack.Success {
		t.Error("unknown page reported as a success")
	}

	if !pack.Fatal {
		t.Error("unknown page not reported as fatal")
	}
}
