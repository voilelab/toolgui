package tgwails

import (
	"encoding/base64"
	"encoding/json"
	"testing"
	"time"

	"github.com/voilelab/toolgui/toolgui/tgframe"
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
	if json.Unmarshal([]byte(packJSON), &pack) != nil {
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
	err = json.Unmarshal([]byte(confJSON), &conf)
	if err != nil {
		t.Fatalf("unmarshal app conf: %v", err)
	}

	if len(conf.PageNames) != 1 || conf.PageNames[0] != testPageName {
		t.Fatalf("unexpected page names: %v", conf.PageNames)
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

	if backend.UploadFile("a.txt", "") != ErrNoSession {
		t.Fatal("expect ErrNoSession from UploadFile before Start")
	}
}

func TestToolGUIUploadFile(t *testing.T) {
	const content = "hello file"

	files := make(chan []byte, 1)
	backend, events := newTestToolGUI(t, newTestApp(func(p *tgframe.Params) error {
		files <- p.State.GetFile("a.txt")
		return nil
	}))
	defer backend.shutdown(t.Context())

	err := backend.Start(testPageName)
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	events.waitResult(t)
	<-files

	err = backend.UploadFile("a.txt", base64.StdEncoding.EncodeToString([]byte(content)))
	if err != nil {
		t.Fatalf("UploadFile: %v", err)
	}

	err = backend.Update(`{"type":"input","id":"a.txt","value":"a.txt"}`)
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	events.waitResult(t)

	if got := string(<-files); got != content {
		t.Fatalf("expect the uploaded file in the state, got %q", got)
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

	if backend.UploadFile("a.txt", "not base64!") == nil {
		t.Fatal("expect an error for invalid base64")
	}
}

// TestToolGUIStartSwitchesPage checks a second Start drops the old session's
// state, the way loading another page in the browser does.
func TestToolGUIStartSwitchesPage(t *testing.T) {
	values := make(chan string, 2)
	app := tgframe.NewApp()
	runFunc := func(p *tgframe.Params) error {
		value := p.State.GetString("field")
		if value == nil {
			values <- ""
			return nil
		}
		values <- *value
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
