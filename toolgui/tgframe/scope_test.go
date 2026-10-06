package tgframe_test

import (
	"testing"

	"github.com/voilelab/toolgui/toolgui/tgcomp"
	"github.com/voilelab/toolgui/toolgui/tgframe"
	"github.com/voilelab/toolgui/toolgui/tgtest"
)

func openPage(t *testing.T, f tgframe.RunFunc) *tgtest.Page {
	t.Helper()

	app := tgframe.NewApp()
	app.AddPage("index", "Index", f)

	p := tgtest.Open(t, app, "index")
	if err := p.Err(); err != nil {
		t.Fatalf("Err = %v", err)
	}

	return p
}

// greeter is a helper written once and called more than once on a page.
func greeter(c *tgframe.Container, side string) {
	c = c.Scope(side)

	name := tgcomp.Textbox(c, "Name")
	if tgcomp.Button(c, "Greet") {
		tgcomp.Text(c, side+" greets "+name)
	}
}

func TestScopeHelperTwice(t *testing.T) {
	p := openPage(t, func(p *tgframe.Params) error {
		greeter(p.Main, "left")
		greeter(p.Main, "right")
		return nil
	})

	p.Get(tgframe.ScopedID("textbox_component_Name", "left")).Input("Alice")
	p.Get(tgframe.ScopedID("textbox_component_Name", "right")).Input("Bob")

	p.Get(tgframe.ScopedID("button_component_Greet", "right")).Click()
	if !p.HasText("right greets Bob") {
		t.Error("expect right greets Bob")
	}
	if p.HasText("left greets") {
		t.Error("left greeted on right's click")
	}
}

func TestScopeKeepsUnscopedIDs(t *testing.T) {
	p := openPage(t, func(p *tgframe.Params) error {
		tgcomp.Button(p.Main, "Save")
		tgcomp.Button(p.Main, "Save", &tgcomp.ButtonConf{ID: "all"})
		return nil
	})

	p.Get("button_component_Save")
	p.Get("button_component_all")
}

func TestScopeWritesInPlace(t *testing.T) {
	p := openPage(t, func(p *tgframe.Params) error {
		tgcomp.Text(p.Main, "one")
		tgcomp.Text(p.Main.Scope("s"), "two")
		tgcomp.Text(p.Main, "three")
		return nil
	})

	children := p.Main().Children
	if len(children) != 3 {
		t.Fatalf("main has %d children, want 3", len(children))
	}

	for i, want := range []string{"one", "two", "three"} {
		if got := children[i].String("text"); got != want {
			t.Errorf("child %d = %q, want %q", i, got, want)
		}
	}
}

func TestScopeNestsAndReachesChildContainers(t *testing.T) {
	p := openPage(t, func(p *tgframe.Params) error {
		c := p.Main.Scope("a").Scope("b")

		box := tgcomp.Box(c, &tgcomp.BoxConf{ID: "box"})
		tgcomp.Button(box, "Go", &tgcomp.ButtonConf{ID: "go"})

		left, _ := tgcomp.Column2(c)
		tgcomp.Button(left, "Go")
		return nil
	})

	p.Get(tgframe.ScopedID("box_component_box", "a", "b"))
	p.Get(tgframe.ScopedID("button_component_go", "a", "b"))
	p.Get(tgframe.ScopedID("button_component_Go", "a", "b"))
}

func TestScopeButtonClicked(t *testing.T) {
	p := openPage(t, func(p *tgframe.Params) error {
		c := p.Main.Scope("s")

		// Asked before the button is drawn.
		if tgcomp.ButtonClicked(c, "Load") {
			tgcomp.Text(c, "loaded")
		}
		tgcomp.Button(c, "Load")
		return nil
	})

	p.Get(tgframe.ScopedID("button_component_Load", "s")).Click()
	if !p.HasText("loaded") {
		t.Error("expect ButtonClicked to see the scoped click")
	}
}

func TestScopeSlotClearReleasesScopedIDs(t *testing.T) {
	show := true
	p := openPage(t, func(p *tgframe.Params) error {
		slot := tgcomp.Empty(p.Main.Scope("s"))

		// Rewriting the same widget claims its scoped id again.
		for range 2 {
			slot.With(func(c *tgframe.Container) {
				tgcomp.Textbox(c, "Name")
			})
		}

		if !show {
			slot.Clear()
		}
		return nil
	})

	id := tgframe.ScopedID("textbox_component_Name", "s")
	p.Get(id).Input("Alice")
	if v, _ := p.State().Get[string](id); v != "Alice" {
		t.Fatalf("state = %q, want Alice", v)
	}

	show = false
	p.Rerun()
	if err := p.Err(); err != nil {
		t.Fatalf("Err = %v", err)
	}

	if _, ok := p.State().Get[string](id); ok {
		t.Error("expect the cleared scoped id's state dropped")
	}
}

func TestScopedIDUnambiguous(t *testing.T) {
	ids := []string{
		tgframe.ScopedID("x", "a/b"),
		tgframe.ScopedID("x", "a", "b"),
		tgframe.ScopedID("b/x", "a"),
		tgframe.ScopedID("x", "a%2Fb"),
		"a/b/x",
	}

	seen := map[string]int{}
	for i, id := range ids {
		if j, ok := seen[id]; ok {
			t.Errorf("ids %d and %d are both %q", j, i, id)
		}
		seen[id] = i
	}

	if got := tgframe.ScopedID("x"); got != "x" {
		t.Errorf("ScopedID without scope = %q, want x", got)
	}
}
