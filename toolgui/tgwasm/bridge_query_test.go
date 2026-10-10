//go:build js && wasm

package tgwasm

import (
	"strings"
	"syscall/js"
	"testing"
	"time"

	"github.com/voilelab/toolgui/toolgui/tgframe"
)

// TestStartPassesQuery checks the second argument of start is what the page
// reads as its query.
func TestStartPassesQuery(t *testing.T) {
	got := make(chan string, 1)
	app := tgframe.NewApp()
	app.AddPage("detail", "Detail", func(p *tgframe.Params) error {
		got <- p.Query.Get("group")
		return nil
	})

	b := newBridge(app)
	b.jsStart(js.Undefined(), []js.Value{
		js.ValueOf("detail"), js.ValueOf("group=a"),
	})
	defer stopped(b)()

	select {
	case g := <-got:
		if g != "a" {
			t.Errorf("group = %q, want a", g)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("the page never ran")
	}
}

// TestStartOversizedQueryOpensNoSession checks a query over the cap is turned
// away like an unknown page.
func TestStartOversizedQueryOpensNoSession(t *testing.T) {
	b := newBridge(testApp())

	query := "x=" + strings.Repeat("a", tgframe.MaxQuerySize)
	b.jsStart(js.Undefined(), []js.Value{js.ValueOf("index"), js.ValueOf(query)})
	settle(b)

	if hostState(b) != nil || b.host.Session() != nil {
		t.Error("expect no session for an oversized query")
	}
}
