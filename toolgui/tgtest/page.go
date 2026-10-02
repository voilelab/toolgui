// Package tgtest drives a toolgui page from a Go test, the way a browser
// would but without one: render the page, then click, type, select and
// upload, and read back what it draws.
//
//	func TestHello(t *testing.T) {
//		p := tgtest.Open(t, newApp(), "index")
//
//		p.GetByLabel("Name").Input("Alice")
//		p.GetByLabel("Greet").Click()
//
//		if err := p.Err(); err != nil {
//			t.Fatal(err)
//		}
//
//		if !p.HasText("Hello, Alice") {
//			t.Error("no greeting")
//		}
//	}
package tgtest

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/voilelab/toolgui/toolgui/tgframe"
	"github.com/voilelab/toolgui/toolgui/tgjson"
)

// DefaultTimeout is how long an action waits for its run to finish.
const DefaultTimeout = 10 * time.Second

// Page is one open page of an app, with a state of its own.
//
// Its methods are meant for the test goroutine. Each action waits for the run
// it starts, so the tree read after it is the one that run drew.
type Page struct {
	t       testing.TB
	state   *tgframe.State
	session *tgframe.Session

	mainID    string
	sidebarID string

	// mu guards forest and lastErr, which the run goroutine writes.
	mu      sync.Mutex
	forest  *forest
	lastErr error

	results chan *tgframe.ResultPack
	result  *tgframe.ResultPack

	main    *Node
	sidebar *Node

	// forms holds the events written inside a form until it is submitted,
	// by the form's key.
	forms map[string][]tgframe.Event

	timeout time.Duration
}

// Open opens page pageName of app and runs it once. The page is closed when
// the test ends.
func Open(t testing.TB, app *tgframe.App, pageName string) *Page {
	t.Helper()

	conf := app.AppConf()
	p := &Page{
		t:         t,
		state:     tgframe.NewState(),
		mainID:    conf.MainContainerID,
		sidebarID: conf.SidebarContainerID,
		forest:    newForest(conf.MainContainerID, conf.SidebarContainerID),
		results:   make(chan *tgframe.ResultPack, 16),
		forms:     map[string][]tgframe.Event{},
		timeout:   DefaultTimeout,
	}

	session, err := tgframe.NewSession(app, pageName, p.state, p.receive)
	if err != nil {
		p.state.Destroy()
		t.Fatalf("tgtest: open page %q: %v", pageName, err)
	}
	p.session = session

	t.Cleanup(func() {
		session.Close()
		p.state.Destroy()
	})

	p.Rerun()
	return p
}

// SetTimeout sets how long an action waits for its run, [DefaultTimeout] by
// default.
func (p *Page) SetTimeout(d time.Duration) {
	p.timeout = d
}

// State is the state the page runs on, for what the actions don't cover.
func (p *Page) State() *tgframe.State {
	return p.state
}

// Result is the result pack of the last run.
func (p *Page) Result() *tgframe.ResultPack {
	return p.result
}

// Err is the error the last run reported, nil if it succeeded.
func (p *Page) Err() error {
	if p.result == nil || p.result.Success {
		return nil
	}

	return errors.New(p.result.Error)
}

// Rerun runs the page again with no change, like the rerun button.
func (p *Page) Rerun() {
	p.t.Helper()
	p.Send(&tgframe.EventEmpty{})
}

// Send applies event and waits for the run it starts.
func (p *Page) Send(event tgframe.Event) {
	p.t.Helper()

	p.session.HandleEvent(event)
	p.wait()
}

// receive is the session's [tgframe.SendPackFunc]. Packs go through json and
// are routed like the web client does, so the tree is what a browser sees.
func (p *Page) receive(pack any) error {
	bs, err := tgjson.Marshal(pack)
	if err != nil {
		return err
	}

	var head struct {
		Success *bool `json:"success"`
		Ready   *bool `json:"ready"`
	}
	if err := tgjson.Unmarshal(bs, &head); err != nil {
		return err
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	switch {
	case head.Success != nil:
		var result tgframe.ResultPack
		if err := tgjson.Unmarshal(bs, &result); err != nil {
			return err
		}

		p.forest.endRun(result.Success)
		p.results <- &result

	case head.Ready != nil:
		p.forest.beginRun()

	default:
		var notify notifyPack
		if err := tgjson.Unmarshal(bs, &notify); err != nil {
			p.lastErr = fmt.Errorf("decode notify pack: %w", err)
			return err
		}

		p.forest.apply(&notify)
	}

	return nil
}

func (p *Page) wait() {
	p.t.Helper()

	timer := time.NewTimer(p.timeout)
	defer timer.Stop()

	select {
	case result := <-p.results:
		p.result = result
	case <-timer.C:
		p.t.Fatalf("tgtest: run did not finish in %v", p.timeout)
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	if p.lastErr != nil {
		err := p.lastErr
		p.lastErr = nil
		p.t.Fatalf("tgtest: %v", err)
	}

	p.main = p.forest.snapshot(p, p.mainID, nil)
	p.sidebar = p.forest.snapshot(p, p.sidebarID, nil)

	// A form gone from the page takes what it held with it.
	mounted := map[string]bool{}
	for _, form := range p.FindByName(formComponentName) {
		mounted[form.queueKey()] = true
	}
	for key := range p.forms {
		if !mounted[key] {
			delete(p.forms, key)
		}
	}
}

// Main is the main container as the last run left it.
func (p *Page) Main() *Node {
	return p.main
}

// Sidebar is the sidebar container as the last run left it.
func (p *Page) Sidebar() *Node {
	return p.sidebar
}

// Find returns every node, main then sidebar, that match accepts.
func (p *Page) Find(match func(n *Node) bool) []*Node {
	var found []*Node
	for _, root := range []*Node{p.main, p.sidebar} {
		root.walk(func(n *Node) {
			if match(n) {
				found = append(found, n)
			}
		})
	}

	return found
}

// FindByName returns every node of component type name, e.g.
// "button_component".
func (p *Page) FindByName(name string) []*Node {
	return p.Find(func(n *Node) bool { return n.Name == name })
}

// Get returns the node of component id, failing the test without one.
func (p *Page) Get(id string) *Node {
	p.t.Helper()

	return p.one(fmt.Sprintf("id %q", id),
		p.Find(func(n *Node) bool { return n.ID == id }))
}

// GetByLabel returns the node labelled label, failing the test unless there
// is exactly one.
func (p *Page) GetByLabel(label string) *Node {
	p.t.Helper()

	return p.one(fmt.Sprintf("label %q", label),
		p.Find(func(n *Node) bool { return n.String("label") == label }))
}

func (p *Page) one(what string, nodes []*Node) *Node {
	p.t.Helper()

	switch len(nodes) {
	case 0:
		p.t.Fatalf("tgtest: no component of %s", what)
	case 1:
	default:
		p.t.Fatalf("tgtest: %d components of %s", len(nodes), what)
	}

	return nodes[0]
}

// textProps are the props a component draws as text. The rest, e.g. a
// fileupload's accept or a link's url, never reach the screen as text.
var textProps = map[string]bool{
	"text": true, "label": true, "title": true, "body": true,
	"message": true, "placeholder": true, "submit_label": true,
	"items": true, "tabs": true, "head": true, "rows": true, "table": true,
	"value": true, "delta": true, "code": true, "latex": true, "html": true,
	"labels": true, "series": true, "x_label": true, "y_label": true,
}

// nestedTextProps are the fields drawn as text inside a prop: a menu item's
// label, a chart series' name.
var nestedTextProps = map[string]bool{"label": true, "name": true}

// HasText reports whether any text the page draws contains s.
func (p *Page) HasText(s string) bool {
	return len(p.Find(func(n *Node) bool {
		for k, v := range n.Props {
			if textProps[k] && hasText(v, s) {
				return true
			}
		}
		return false
	})) > 0
}

func hasText(v any, s string) bool {
	switch v := v.(type) {
	case string:
		return strings.Contains(v, s)
	case []any:
		for _, e := range v {
			if hasText(e, s) {
				return true
			}
		}
	case map[string]any:
		for k, e := range v {
			if nestedTextProps[k] && hasText(e, s) {
				return true
			}
		}
	}

	return false
}
