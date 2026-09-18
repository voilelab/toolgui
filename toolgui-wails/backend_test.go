package tgwails

import (
	"encoding/base64"
	"sync"
	"testing"
	"time"

	"github.com/voilelab/toolgui/toolgui/tgframe"
	"github.com/voilelab/toolgui/toolgui/tgjson"
)

const testPageName = "index"

// fakeEvents stands in for the window's event stream, recording the packs the
// frontend would receive.
type fakeEvents struct {
	packs chan map[string]any
}

func newFakeEvents() *fakeEvents {
	return &fakeEvents{packs: make(chan map[string]any, 256)}
}

func (f *fakeEvents) emit(packJSON string) {
	var pack map[string]any
	if tgjson.Unmarshal([]byte(packJSON), &pack) != nil {
		return
	}

	f.packs <- pack
}

// waitResult collects packs until the run reports its result.
func (f *fakeEvents) waitResult(t *testing.T) (map[string]any, []map[string]any) {
	t.Helper()

	var seen []map[string]any
	for {
		select {
		case pack := <-f.packs:
			if _, isResult := pack["success"]; isResult {
				return pack, seen
			}
			seen = append(seen, pack)
		case <-time.After(5 * time.Second):
			t.Fatal("timed out waiting for a result pack")
		}
	}
}

func newTestToolGUI(t *testing.T, app *tgframe.App) (*ToolGUI, *fakeEvents) {
	t.Helper()

	events := newFakeEvents()
	backend := NewToolGUI(app)
	backend.emit = events.emit

	return backend, events
}

func newTestApp(runFunc tgframe.RunFunc) *tgframe.App {
	app := tgframe.NewApp()
	app.AddPage(testPageName, "Index", runFunc)
	return app
}

func addTestComponent(p *tgframe.Params, id string) {
	p.Main.AddComponent(&tgframe.BaseComponent{Name: "test_component", ID: id})
}

func TestToolGUIAppConf(t *testing.T) {
	backend, _ := newTestToolGUI(t, newTestApp(func(p *tgframe.Params) error {
		return nil
	}))

	confJSON, err := backend.AppConf()
	if err != nil {
		t.Fatalf("AppConf: %v", err)
	}

	var conf tgframe.AppConf
	err = tgjson.Unmarshal([]byte(confJSON), &conf)
	if err != nil {
		t.Fatalf("unmarshal app conf: %v", err)
	}

	if len(conf.PageNames) != 1 || conf.PageNames[0] != testPageName {
		t.Fatalf("unexpected page names: %v", conf.PageNames)
	}
}

// TestToolGUIAppConfMenu pins where the desktop conf differs from the web
// one: the window draws the menu itself, so the tree is kept out of what the
// frontend is told and no second menubar appears inside the window.
func TestToolGUIAppConfMenu(t *testing.T) {
	app := newTestApp(func(p *tgframe.Params) error { return nil })
	app.SetMenu(testMenu())

	backend, _ := newTestToolGUI(t, app)

	confJSON, err := backend.AppConf()
	if err != nil {
		t.Fatalf("AppConf: %v", err)
	}

	var conf tgframe.AppConf
	if err := tgjson.Unmarshal([]byte(confJSON), &conf); err != nil {
		t.Fatalf("unmarshal app conf: %v", err)
	}

	if conf.Menu != nil {
		t.Fatalf("AppConf carries a menu the window already draws: %v",
			conf.Menu)
	}

	// The rest of the conf is untouched, so dropping the menu is not done by
	// handing the frontend a hollowed out config.
	if len(conf.PageNames) != 1 || conf.PageNames[0] != testPageName {
		t.Fatalf("unexpected page names: %v", conf.PageNames)
	}

	// And the App still has its menu: the conf is a copy, so the tree the
	// window was built from is not what was emptied.
	if len(app.AppConf().Menu) != 1 {
		t.Fatalf("AppConf dropped the App's own menu: %v", app.AppConf().Menu)
	}
}

func TestToolGUIStartRunsPage(t *testing.T) {
	backend, events := newTestToolGUI(t, newTestApp(func(p *tgframe.Params) error {
		addTestComponent(p, "comp")
		return nil
	}))
	defer backend.shutdown(t.Context())

	err := backend.Start(testPageName)
	if err != nil {
		t.Fatalf("Start: %v", err)
	}

	result, seen := events.waitResult(t)
	if result["success"] != true {
		t.Fatalf("expect a successful run, got %v", result)
	}

	// ready pack then the create pack for the component
	if len(seen) != 2 {
		t.Fatalf("expect 2 packs before the result, got %v", seen)
	}
	if seen[0]["ready"] != true {
		t.Fatalf("expect a ready pack first, got %v", seen[0])
	}
}

func TestToolGUIStartUnknownPage(t *testing.T) {
	backend, _ := newTestToolGUI(t, newTestApp(func(p *tgframe.Params) error {
		return nil
	}))

	if backend.Start("nope") == nil {
		t.Fatal("expect an error for an unknown page")
	}
}

func TestToolGUIUpdateReruns(t *testing.T) {
	backend, events := newTestToolGUI(t, newTestApp(func(p *tgframe.Params) error {
		clicked := p.State.GetClickID() == "click_me"
		if clicked {
			addTestComponent(p, "clicked")
		}
		return nil
	}))
	defer backend.shutdown(t.Context())

	err := backend.Start(testPageName)
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	events.waitResult(t)

	err = backend.Update(`{"type":"click","id":"click_me"}`)
	if err != nil {
		t.Fatalf("Update: %v", err)
	}

	result, seen := events.waitResult(t)
	if result["success"] != true {
		t.Fatalf("expect a successful rerun, got %v", result)
	}
	if len(seen) != 2 {
		t.Fatalf("expect the click to add a component, got %v", seen)
	}
}

func TestToolGUIUpdateBadEvent(t *testing.T) {
	backend, _ := newTestToolGUI(t, newTestApp(func(p *tgframe.Params) error {
		return nil
	}))
	defer backend.shutdown(t.Context())

	err := backend.Start(testPageName)
	if err != nil {
		t.Fatalf("Start: %v", err)
	}

	if backend.Update(`{"type":"nope"}`) == nil {
		t.Fatal("expect an error for an unknown event type")
	}
}

func TestToolGUIBeforeStart(t *testing.T) {
	backend, _ := newTestToolGUI(t, newTestApp(func(p *tgframe.Params) error {
		return nil
	}))

	if backend.Update(`{"type":"click","id":"x"}`) != ErrNoSession {
		t.Fatal("expect ErrNoSession from Update before Start")
	}

	if _, err := backend.UploadFileStart("a.txt"); err != ErrNoSession {
		t.Fatal("expect ErrNoSession from UploadFileStart before Start")
	}

	if backend.UploadFileFinish("f", "1") != ErrNoSession {
		t.Fatal("expect ErrNoSession from UploadFileFinish before Start")
	}
}

// TestToolGUIClickMenu is the way back from the native menubar: the callback
// happens in Go, so the click goes straight into the session and the page
// reruns with tgframe.MenuClicked seeing it.
func TestToolGUIClickMenu(t *testing.T) {
	app := newTestApp(func(p *tgframe.Params) error {
		if tgframe.MenuClicked(p, "file_open") {
			addTestComponent(p, "opened")
		}
		return nil
	})
	app.SetMenu(testMenu())

	backend, events := newTestToolGUI(t, app)
	defer backend.shutdown(t.Context())

	if err := backend.Start(testPageName); err != nil {
		t.Fatalf("Start: %v", err)
	}
	events.waitResult(t)

	backend.clickMenu(tgframe.MenuID("file_open"))

	result, seen := events.waitResult(t)
	if result["success"] != true {
		t.Fatalf("expect a successful rerun, got %v", result)
	}
	if len(seen) != 2 {
		t.Fatalf("expect the click to add a component, got %v", seen)
	}
}

// TestToolGUIClickMenuBeforeStart is the window between the menubar appearing
// and the first page: the menu is already there to be picked from, and there
// is no session behind it yet.
func TestToolGUIClickMenuBeforeStart(t *testing.T) {
	app := newTestApp(func(p *tgframe.Params) error { return nil })
	app.SetMenu(testMenu())

	backend, events := newTestToolGUI(t, app)

	// No panic, and nothing reaches the frontend: there is no run to send
	// packs from.
	backend.clickMenu(tgframe.MenuID("file_open"))

	select {
	case pack := <-events.packs:
		t.Fatalf("a click before Start produced a pack: %v", pack)
	default:
	}
}

// TestToolGUIQueueMenuClickOrder is why the picks go through a queue rather
// than a goroutine each: a run cuts the one before it, so two picks that
// overtake each other would leave the page showing the older one.
func TestToolGUIQueueMenuClickOrder(t *testing.T) {
	var lock sync.Mutex
	var seen []string

	app := newTestApp(func(p *tgframe.Params) error {
		if id := p.State.GetClickID(); id != "" {
			lock.Lock()
			seen = append(seen, id)
			lock.Unlock()
		}
		return nil
	})
	app.SetMenu(testMenu())

	backend, events := newTestToolGUI(t, app)
	defer backend.shutdown(t.Context())

	if err := backend.Start(testPageName); err != nil {
		t.Fatalf("Start: %v", err)
	}
	events.waitResult(t)

	// The opening run drew nothing under a click, so the log starts empty.
	lock.Lock()
	seen = nil
	lock.Unlock()

	// Queued back to back, the way the message loop delivers them.
	want := []string{
		tgframe.MenuID("file_open"),
		tgframe.MenuID("file_reload"),
		tgframe.MenuID("file_quit"),
	}
	for _, id := range want {
		backend.queueMenuClick(id)
	}

	// Waiting on the page func rather than on result packs: each pick cuts
	// the run before it, and a cut run sends no result, so three picks in a
	// row are not three results. What they are is three runs, in order.
	count := func() int {
		lock.Lock()
		defer lock.Unlock()

		return len(seen)
	}

	deadline := time.Now().Add(5 * time.Second)
	for count() < len(want) && time.Now().Before(deadline) {
		time.Sleep(time.Millisecond)
	}

	lock.Lock()
	defer lock.Unlock()

	if len(seen) != len(want) {
		t.Fatalf("runs = %v, want one per pick (%v)", seen, want)
	}
	for i := range want {
		if seen[i] != want[i] {
			t.Errorf("run %d handled %q, want %q", i, seen[i], want[i])
		}
	}
}

// uploadFile sends content the way the frontend does: start, chunk, finish.
func uploadFile(t *testing.T, backend *ToolGUI, componentID, name string, chunks ...string) {
	t.Helper()

	uploadID, err := backend.UploadFileStart(name)
	if err != nil {
		t.Fatalf("UploadFileStart: %v", err)
	}

	for _, chunk := range chunks {
		err = backend.UploadFileChunk(uploadID,
			base64.StdEncoding.EncodeToString([]byte(chunk)))
		if err != nil {
			t.Fatalf("UploadFileChunk: %v", err)
		}
	}

	if err := backend.UploadFileFinish(componentID, uploadID); err != nil {
		t.Fatalf("UploadFileFinish: %v", err)
	}
}

// TestToolGUIUploadFileChunk checks a file that crosses the bridge in pieces
// lands whole, in the order the chunks were sent.
func TestToolGUIUploadFileChunk(t *testing.T) {
	const componentID = "fileupload_component_file"

	files := make(chan []byte, 1)
	backend, events := newTestToolGUI(t, newTestApp(func(p *tgframe.Params) error {
		// A real fileupload draws itself under this id, and the session only
		// takes an event naming a component the page is showing.
		addTestComponent(p, componentID)

		file := p.State.GetFile(componentID)
		if file == nil {
			files <- nil
			return nil
		}

		bs, err := file.Bytes()
		if err != nil {
			return err
		}

		files <- bs
		return nil
	}))
	defer backend.shutdown(t.Context())

	err := backend.Start(testPageName)
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	events.waitResult(t)
	<-files

	uploadFile(t, backend, componentID, "a.txt", "hello ", "file")

	err = backend.Update(`{"type":"input","id":"` + componentID + `","value":"a.txt"}`)
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	events.waitResult(t)

	if got := string(<-files); got != "hello file" {
		t.Fatalf("expect the uploaded file in the state, got %q", got)
	}
}

// TestToolGUIUploadFileHidesUntilFinish checks the component sees nothing
// until the last chunk lands, so a page that reruns mid-upload doesn't read
// half a file.
func TestToolGUIUploadFileHidesUntilFinish(t *testing.T) {
	const componentID = "fileupload_component_file"

	backend, events := newTestToolGUI(t, newTestApp(func(p *tgframe.Params) error {
		return nil
	}))
	defer backend.shutdown(t.Context())

	if err := backend.Start(testPageName); err != nil {
		t.Fatalf("Start: %v", err)
	}
	events.waitResult(t)

	uploadID, err := backend.UploadFileStart("a.txt")
	if err != nil {
		t.Fatalf("UploadFileStart: %v", err)
	}

	err = backend.UploadFileChunk(uploadID,
		base64.StdEncoding.EncodeToString([]byte("half")))
	if err != nil {
		t.Fatalf("UploadFileChunk: %v", err)
	}

	if backend.state.GetFile(componentID) != nil {
		t.Error("expect no file under the component while it is still arriving")
	}

	if err := backend.UploadFileFinish(componentID, uploadID); err != nil {
		t.Fatalf("UploadFileFinish: %v", err)
	}

	if backend.state.GetFile(componentID) == nil {
		t.Error("expect the file under the component once it finished")
	}
}

// TestToolGUIUploadFileOverlapping checks two picks racing on one component
// keep their own content: the second replaces the first, whatever order the
// chunks arrive in.
func TestToolGUIUploadFileOverlapping(t *testing.T) {
	const componentID = "fileupload_component_file"

	backend, events := newTestToolGUI(t, newTestApp(func(p *tgframe.Params) error {
		return nil
	}))
	defer backend.shutdown(t.Context())

	if err := backend.Start(testPageName); err != nil {
		t.Fatalf("Start: %v", err)
	}
	events.waitResult(t)

	firstID, err := backend.UploadFileStart("first.txt")
	if err != nil {
		t.Fatalf("UploadFileStart: %v", err)
	}

	secondID, err := backend.UploadFileStart("second.txt")
	if err != nil {
		t.Fatalf("UploadFileStart: %v", err)
	}

	chunk := func(uploadID, content string) {
		t.Helper()

		err := backend.UploadFileChunk(uploadID,
			base64.StdEncoding.EncodeToString([]byte(content)))
		if err != nil {
			t.Fatalf("UploadFileChunk: %v", err)
		}
	}

	chunk(firstID, "one")
	chunk(secondID, "two")
	chunk(firstID, "one")
	chunk(secondID, "two")

	if err := backend.UploadFileFinish(componentID, firstID); err != nil {
		t.Fatalf("UploadFileFinish: %v", err)
	}

	if err := backend.UploadFileFinish(componentID, secondID); err != nil {
		t.Fatalf("UploadFileFinish: %v", err)
	}

	file := backend.state.GetFile(componentID)
	if file == nil {
		t.Fatal("expect a file under the component")
	}

	bs, err := file.Bytes()
	if err != nil {
		t.Fatalf("Bytes: %v", err)
	}

	if string(bs) != "twotwo" {
		t.Errorf("expect the second upload's own content, got %q", bs)
	}

	if file.Name() != "second.txt" {
		t.Errorf("Name = %q, want second.txt", file.Name())
	}
}

// TestToolGUIUploadFileUnknownID checks a chunk for an upload the session
// doesn't know, which is what one sent after a session switch looks like, is
// refused rather than written somewhere.
func TestToolGUIUploadFileUnknownID(t *testing.T) {
	backend, events := newTestToolGUI(t, newTestApp(func(p *tgframe.Params) error {
		return nil
	}))
	defer backend.shutdown(t.Context())

	if err := backend.Start(testPageName); err != nil {
		t.Fatalf("Start: %v", err)
	}
	events.waitResult(t)

	err := backend.UploadFileChunk("no_such_upload",
		base64.StdEncoding.EncodeToString([]byte("x")))
	if err != ErrNoUpload {
		t.Errorf("UploadFileChunk error = %v, want ErrNoUpload", err)
	}

	if err := backend.UploadFileFinish("comp", "no_such_upload"); err != ErrNoUpload {
		t.Errorf("UploadFileFinish error = %v, want ErrNoUpload", err)
	}
}

// TestToolGUIStartDropsPendingUploads checks a session switch drops the
// uploads the old session had in flight.
func TestToolGUIStartDropsPendingUploads(t *testing.T) {
	backend, events := newTestToolGUI(t, newTestApp(func(p *tgframe.Params) error {
		return nil
	}))
	defer backend.shutdown(t.Context())

	if err := backend.Start(testPageName); err != nil {
		t.Fatalf("Start: %v", err)
	}
	events.waitResult(t)

	uploadID, err := backend.UploadFileStart("a.txt")
	if err != nil {
		t.Fatalf("UploadFileStart: %v", err)
	}

	if err := backend.Start(testPageName); err != nil {
		t.Fatalf("Start: %v", err)
	}
	events.waitResult(t)

	err = backend.UploadFileChunk(uploadID,
		base64.StdEncoding.EncodeToString([]byte("x")))
	if err != ErrNoUpload {
		t.Errorf("UploadFileChunk error = %v, want ErrNoUpload", err)
	}
}

func TestToolGUIUploadFileBadBase64(t *testing.T) {
	backend, events := newTestToolGUI(t, newTestApp(func(p *tgframe.Params) error {
		return nil
	}))
	defer backend.shutdown(t.Context())

	err := backend.Start(testPageName)
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	events.waitResult(t)

	uploadID, err := backend.UploadFileStart("a.txt")
	if err != nil {
		t.Fatalf("UploadFileStart: %v", err)
	}

	if backend.UploadFileChunk(uploadID, "not base64!") == nil {
		t.Fatal("expect an error for invalid base64")
	}
}

// TestToolGUIStartSwitchesPage checks a second Start drops the old session's
// state, the way loading another page in the browser does.
func TestToolGUIStartSwitchesPage(t *testing.T) {
	values := make(chan string, 2)
	app := tgframe.NewApp()
	runFunc := func(p *tgframe.Params) error {
		// The input event below names this component, so the page has to be
		// showing it, the way the component reading the value would draw it.
		addTestComponent(p, "field")

		value, _ := p.State.Get[string]("field")
		values <- value
		return nil
	}
	app.AddPage("first", "First", runFunc)
	app.AddPage("second", "Second", runFunc)

	backend, events := newTestToolGUI(t, app)
	defer backend.shutdown(t.Context())

	err := backend.Start("first")
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	events.waitResult(t)
	<-values

	err = backend.Update(`{"type":"input","id":"field","value":"typed"}`)
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	events.waitResult(t)
	if got := <-values; got != "typed" {
		t.Fatalf("expect the value in the state, got %q", got)
	}

	err = backend.Start("second")
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	events.waitResult(t)
	if got := <-values; got != "" {
		t.Fatalf("expect a fresh state on the new page, got %q", got)
	}
}
