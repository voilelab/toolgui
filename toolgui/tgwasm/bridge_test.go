//go:build js && wasm

package tgwasm

import (
	"encoding/json"
	"strings"
	"syscall/js"
	"testing"
	"time"

	"github.com/voilelab/toolgui/toolgui/tgframe"
	"github.com/voilelab/toolgui/toolgui/tgutil"
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

// slot is jsNewUpload's answer, read back the way the page reads it.
type slot struct {
	Dir   []string `json:"dir"`
	Name  string   `json:"name"`
	Error string   `json:"error"`
}

// newUpload asks the bridge for somewhere to write, the way the worker does.
//
// It retries, which the worker does not. The state's directory is a chain of
// promises away when a session opens, and reserving a file says so rather than
// waiting -- an upload arrives where waiting stops the event loop the promises
// need. A user cannot pick a file in the turns of it that takes; a test that
// uploads in the same breath as starting the session can.
func newUpload(t *testing.T, b *bridge) slot {
	t.Helper()

	last := ""

	for range 400 {
		// A fresh one each time: the answer leaves out what it has nothing to
		// say about, and unmarshalling into a used one would keep the last
		// error rather than clear it.
		var got slot

		if err := json.Unmarshal([]byte(b.jsNewUpload(js.Undefined(), nil).(string)), &got); err != nil {
			t.Fatalf("read the upload slot: %v", err)
		}

		if got.Error == "" {
			return got
		}

		last = got.Error

		time.Sleep(5 * time.Millisecond)
	}

	t.Fatalf("newUpload: %s", last)
	return slot{}
}

// writeUpload does the page's half: it walks to the reserved file, streams a
// blob into it, closes the stream and opens the sync access handle the bridge
// is handed. Everything awaited is awaited here, on this side of the boundary.
func writeUpload(t *testing.T, s slot, content string) js.Value {
	t.Helper()

	dir, err := await(js.Global().Get("navigator").Get("storage").Call("getDirectory"))
	if err != nil {
		t.Fatalf("open the origin private file system: %v", err)
	}

	for _, name := range s.Dir {
		dir, err = await(dir.Call("getDirectoryHandle", name))
		if err != nil {
			t.Fatalf("open %q: %v", name, err)
		}
	}

	file, err := await(dir.Call("getFileHandle", s.Name,
		map[string]any{"create": true}))
	if err != nil {
		t.Fatalf("open the reserved file: %v", err)
	}

	writable, err := await(file.Call("createWritable",
		map[string]any{"keepExistingData": false}))
	if err != nil {
		t.Fatalf("open a writable stream: %v", err)
	}

	blob := js.Global().Get("Blob").New([]any{content})

	if _, err := await(blob.Call("stream").Call("pipeTo", writable)); err != nil {
		t.Fatalf("pipe the blob into the file: %v", err)
	}

	handle, err := await(file.Call("createSyncAccessHandle"))
	if err != nil {
		t.Fatalf("open a sync access handle: %v", err)
	}

	return handle
}

// await settles a promise. Test code runs on a goroutine of its own, which is
// the one place in this package where waiting for one is allowed.
func await(p js.Value) (js.Value, error) {
	type settled struct {
		value js.Value
		err   error
	}

	ch := make(chan settled, 1)

	var onValue, onReason js.Func

	release := func() {
		onValue.Release()
		onReason.Release()
	}

	onValue = js.FuncOf(func(_ js.Value, args []js.Value) any {
		ch <- settled{value: args[0]}
		release()
		return nil
	})

	onReason = js.FuncOf(func(_ js.Value, args []js.Value) any {
		ch <- settled{err: jsError(args[0])}
		release()
		return nil
	})

	p.Call("then", onValue, onReason)

	s := <-ch
	return s.value, s.err
}

// jsError is enough of a rejection reason for a test to print.
func jsError(v js.Value) error {
	return tgutil.Errorf("%s: %s", v.Get("name"), v.Get("message"))
}

// stopped closes a bridge's session, so the state's directory and the handles
// in it go rather than piling up for the rest of the test binary's life.
func stopped(b *bridge) func() {
	return func() {
		b.lock.Lock()
		defer b.lock.Unlock()

		b.closeSession()
	}
}

// TestUploadFileStoresWhatThePageWrote checks the handover: the bridge names a
// file, the page writes it, and what the component gets back is what was
// written. No bytes cross the boundary at any point.
func TestUploadFileStoresWhatThePageWrote(t *testing.T) {
	b := newBridge(testApp())
	b.jsStart(js.Undefined(), []js.Value{js.ValueOf("index")})
	defer stopped(b)()

	s := newUpload(t, b)
	handle := writeUpload(t, s, "hello file")

	got := b.jsUploadFile(js.Undefined(), []js.Value{
		js.ValueOf("comp"), js.ValueOf("a.txt"), js.ValueOf(s.Name), handle,
	})
	if got != "" {
		t.Fatalf("uploadFile: %v", got)
	}

	file := b.state.GetFile("comp")
	if file == nil {
		t.Fatal("expect the upload to be stored under the component")
	}

	if file.Name() != "a.txt" {
		t.Errorf("Name = %q, want a.txt", file.Name())
	}

	if file.Size() != 10 {
		t.Errorf("Size = %d, want 10", file.Size())
	}

	bs, err := file.Bytes()
	if err != nil {
		t.Fatalf("Bytes: %v", err)
	}

	if string(bs) != "hello file" {
		t.Errorf("Bytes = %q, want hello file", bs)
	}
}

// TestCancelUploadLeavesTheComponentEmpty checks an upload the page gave up on
// reaches nothing. A file the page could not finish writing is not a file a
// page function should be handed half of.
func TestCancelUploadLeavesTheComponentEmpty(t *testing.T) {
	b := newBridge(testApp())
	b.jsStart(js.Undefined(), []js.Value{js.ValueOf("index")})
	defer stopped(b)()

	s := newUpload(t, b)
	// The worker is holding the file when it gives up, so it lets go first.
	// Everything on a sync access handle is synchronous, close included.
	writeUpload(t, s, "half a f").Call("close")

	b.jsCancelUpload(js.Undefined(), []js.Value{js.ValueOf(s.Name)})

	if b.state.GetFile("comp") != nil {
		t.Error("expect a cancelled upload to reach no component")
	}

	// And the bridge no longer knows the slot, so a late handover is refused
	// rather than storing what was abandoned.
	got := b.jsUploadFile(js.Undefined(), []js.Value{
		js.ValueOf("comp"), js.ValueOf("a.txt"), js.ValueOf(s.Name), js.Undefined(),
	})
	if got != ErrNoUpload.Error() {
		t.Errorf("uploadFile after a cancel = %v, want %v", got, ErrNoUpload)
	}
}

// TestUploadFileNeedsASession checks an upload before start is refused rather
// than stored somewhere nothing will clear up.
func TestUploadFileNeedsASession(t *testing.T) {
	b := newBridge(testApp())

	var got slot
	if err := json.Unmarshal([]byte(b.jsNewUpload(js.Undefined(), nil).(string)), &got); err != nil {
		t.Fatalf("read the upload slot: %v", err)
	}

	if !strings.Contains(got.Error, ErrNoSession.Error()) {
		t.Errorf("newUpload without a session = %q, want %v", got.Error, ErrNoSession)
	}
}

// TestStartDropsUnfinishedUploads checks a page switch lets go of the
// reservations the session before it handed out. Their files went with the
// state's directory, so a handover afterwards must not find a slot to put one
// into -- and the handle that arrives with it has to be closed, because the
// browser will not remove a file anything still holds one for.
func TestStartDropsUnfinishedUploads(t *testing.T) {
	b := newBridge(testApp())
	b.jsStart(js.Undefined(), []js.Value{js.ValueOf("index")})
	defer stopped(b)()

	s := newUpload(t, b)
	handle := writeUpload(t, s, "hello file")

	b.jsStart(js.Undefined(), []js.Value{js.ValueOf("other")})

	if len(b.uploads) != 0 {
		t.Errorf("%d uploads survived the page switch, want none", len(b.uploads))
	}

	got := b.jsUploadFile(js.Undefined(), []js.Value{
		js.ValueOf("comp"), js.ValueOf("a.txt"), js.ValueOf(s.Name), handle,
	})
	if got != ErrNoUpload.Error() {
		t.Errorf("uploadFile after a page switch = %v, want %v", got, ErrNoUpload)
	}

	if handleOpen(handle) {
		t.Error("expect a handle the bridge would not take to be closed")
	}
}

// handleOpen reports whether a sync access handle is still open. A closed one answers
// everything with InvalidStateError.
func handleOpen(handle js.Value) (ok bool) {
	defer func() {
		if recover() != nil {
			ok = false
		}
	}()

	handle.Call("getSize")
	return true
}
