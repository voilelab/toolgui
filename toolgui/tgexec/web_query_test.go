package tgexec

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/voilelab/toolgui/toolgui/tgcomp/tccontent"
	"github.com/voilelab/toolgui/toolgui/tgframe"
	"github.com/voilelab/toolgui/toolgui/tgjson"
	"golang.org/x/net/websocket"
)

// newQueryServer serves a page drawing its query.
func newQueryServer(t *testing.T) *httptest.Server {
	t.Helper()

	app := tgframe.NewApp()
	app.AddPage("detail", "Detail", func(p *tgframe.Params) error {
		tccontent.Text(p.Main, "group="+p.Query.Get("group"))
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

	return srv
}

// The query on the socket url is what the page reads.
func TestUpdateQueryReachesPage(t *testing.T) {
	srv := newQueryServer(t)
	ws := dialUpdate(t, srv, "detail?group=a%20b")

	if err := jsonCodec.Send(ws, stateIDPack{}); err != nil {
		t.Fatalf("send state id: %v", err)
	}

	var pack stateIDPack
	if err := jsonCodec.Receive(ws, &pack); err != nil {
		t.Fatalf("receive: %v", err)
	}

	if err := websocket.Message.Send(ws, []byte(`{}`)); err != nil {
		t.Fatalf("send empty event: %v", err)
	}

	var bs []byte
	if err := websocket.Message.Receive(ws, &bs); err != nil { // ready
		t.Fatalf("receive: %v", err)
	}

	if err := websocket.Message.Receive(ws, &bs); err != nil {
		t.Fatalf("receive: %v", err)
	}

	var notify struct {
		Component map[string]any `json:"component"`
	}
	if err := tgjson.Unmarshal(bs, &notify); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if got := notify.Component["text"]; got != "group=a b" {
		t.Errorf("text = %v, want group=a b", got)
	}
}

// A query over the cap is refused like an unknown page, and its value stays
// out of the log.
func TestUpdateOversizedQueryIsFatal(t *testing.T) {
	logs := captureLog(t)
	srv := newQueryServer(t)

	secret := strings.Repeat("s", tgframe.MaxQuerySize)
	ws := dialUpdate(t, srv, "detail?group="+secret)

	var pack tgframe.ResultPack
	if err := jsonCodec.Receive(ws, &pack); err != nil {
		t.Fatalf("receive: %v", err)
	}

	if pack.Success || !pack.Fatal {
		t.Errorf("pack = %+v, want a fatal error", pack)
	}

	if pack.Error != tgframe.ErrQueryTooLarge.Error() {
		t.Errorf("Error = %q", pack.Error)
	}

	got := waitLog(t, logs, "page query")
	if strings.Contains(got, "sss") {
		t.Error("the log carries the query value")
	}
}

// A query that does not parse is refused too, without echoing it.
func TestUpdateInvalidQueryIsFatal(t *testing.T) {
	logs := captureLog(t)
	srv := newQueryServer(t)
	ws := dialUpdate(t, srv, "detail?group=%zzsecret")

	var pack tgframe.ResultPack
	if err := jsonCodec.Receive(ws, &pack); err != nil {
		t.Fatalf("receive: %v", err)
	}

	if !pack.Fatal || pack.Error != tgframe.ErrInvalidQuery.Error() {
		t.Errorf("pack = %+v, want a fatal invalid query", pack)
	}

	if got := waitLog(t, logs, "page query"); strings.Contains(got, "secret") {
		t.Error("the log carries the query value")
	}
}
