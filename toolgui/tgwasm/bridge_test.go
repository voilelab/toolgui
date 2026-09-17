//go:build js && wasm

package tgwasm

import (
	"syscall/js"
	"testing"

	"github.com/voilelab/toolgui/toolgui/tgframe"
)

func testApp() *tgframe.App {
	app := tgframe.NewApp()

	blank := func(p *tgframe.Params) error { return nil }
	app.AddPage("index", "Index", blank)
	app.AddPage("other", "Other", blank)

	return app
}

// TestStartDestroysTheStateItReplaces checks a page switch lets go of what the
// session before it held. The state's uploads are files in the origin private
// file system with handles open on them, so dropping the pointer leaves them
// there for the life of the tab: nothing else closes them, and the garbage
// collector cannot.
func TestStartDestroysTheStateItReplaces(t *testing.T) {
	b := newBridge(testApp())

	b.jsStart(js.Undefined(), []js.Value{js.ValueOf("index")})

	state := b.state
	if state == nil {
		t.Fatal("expect start to open a session")
	}

	if _, err := state.SetFile("comp", "a.txt", []byte("hello")); err != nil {
		t.Fatalf("SetFile: %v", err)
	}

	b.jsStart(js.Undefined(), []js.Value{js.ValueOf("other")})

	if b.state == state {
		t.Fatal("expect start to open a state of its own")
	}

	if state.GetFile("comp") != nil {
		t.Error("expect the replaced session's state to be destroyed")
	}
}

// TestStartOnAnUnknownPageDestroysTheStateItMade checks the state made for a
// session that never opened is not simply dropped either.
func TestStartOnAnUnknownPageDestroysTheStateItMade(t *testing.T) {
	b := newBridge(testApp())

	b.jsStart(js.Undefined(), []js.Value{js.ValueOf("nowhere")})

	if b.state != nil || b.session != nil {
		t.Error("expect no session after a start on a page that is not there")
	}
}
