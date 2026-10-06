package tgtest_test

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/voilelab/toolgui/toolgui/tgcomp"
	"github.com/voilelab/toolgui/toolgui/tgframe"
	"github.com/voilelab/toolgui/toolgui/tgtest"
)

func newApp(name string, f tgframe.RunFunc) *tgframe.App {
	app := tgframe.NewApp()
	app.AddPage(name, name, f)
	return app
}

func TestRender(t *testing.T) {
	app := newApp("index", func(p *tgframe.Params) error {
		tgcomp.Title(p.Main, "Hello")
		tgcomp.Text(p.Sidebar, "Side")
		return nil
	})

	p := tgtest.Open(t, app, "index")
	if err := p.Err(); err != nil {
		t.Fatalf("Err = %v", err)
	}

	if !p.HasText("Hello") || !p.HasText("Side") {
		t.Error("expect Hello and Side drawn")
	}

	if got := len(p.Main().Children); got != 1 {
		t.Errorf("main has %d children, want 1", got)
	}

	if got := len(p.Sidebar().Children); got != 1 {
		t.Errorf("sidebar has %d children, want 1", got)
	}
}

func TestClickAndInput(t *testing.T) {
	app := newApp("index", func(p *tgframe.Params) error {
		name := tgcomp.Textbox(p.Main, "Name")
		if tgcomp.Button(p.Main, "Greet") {
			tgcomp.Text(p.Main, "Hello, "+name)
		}
		return nil
	})

	p := tgtest.Open(t, app, "index")

	p.GetByLabel("Name").Input("Alice")
	if p.HasText("Hello") {
		t.Error("greeted before the click")
	}

	p.GetByLabel("Greet").Click()
	if !p.HasText("Hello, Alice") {
		t.Error("expect Hello, Alice")
	}

	// A click lasts one run.
	p.Rerun()
	if p.HasText("Hello") {
		t.Error("greeting outlived the click")
	}
}

func TestSelect(t *testing.T) {
	app := newApp("index", func(p *tgframe.Params) error {
		if i := tgcomp.Select(p.Main, "Fruit", []string{"apple", "pear"}); i != nil {
			tgcomp.Text(p.Main, fmt.Sprintf("picked %d", *i))
		}

		picked := tgcomp.MultiSelect(p.Main, "Many", []string{"a", "b", "c"})
		tgcomp.Text(p.Main, fmt.Sprintf("many %v", picked))

		v, _ := tgcomp.Number[int](p.Main, "Count")
		tgcomp.Text(p.Main, fmt.Sprintf("count %d", v))
		return nil
	})

	p := tgtest.Open(t, app, "index")

	p.GetByLabel("Fruit").Select(1)
	p.GetByLabel("Many").SelectMany(0, 2)
	p.GetByLabel("Count").Input(3)

	for _, want := range []string{"picked 1", "many [0 2]", "count 3"} {
		if !p.HasText(want) {
			t.Errorf("expect %q", want)
		}
	}
}

func TestUpload(t *testing.T) {
	app := newApp("index", func(p *tgframe.Params) error {
		if f := tgcomp.FileUpload(p.Main, "File", ".txt"); f != nil {
			bs, err := f.Bytes()
			if err != nil {
				return err
			}
			tgcomp.Text(p.Main, fmt.Sprintf("%s %s %s", f.Name, f.Type, bs))
		}

		files := tgcomp.MultiFileUpload(p.Main, "Files", "")
		for _, f := range files {
			bs, err := f.Bytes()
			if err != nil {
				return err
			}
			tgcomp.Text(p.Main, "multi "+string(bs))
		}
		return nil
	})

	p := tgtest.Open(t, app, "index")

	p.GetByLabel("File").Upload("a.txt", []byte("hello"))
	if err := p.Err(); err != nil {
		t.Fatalf("Err = %v", err)
	}

	if !p.HasText("a.txt text/plain") || !p.HasText("hello") {
		t.Error("expect the uploaded file read back")
	}

	p.GetByLabel("Files").UploadFiles(
		tgtest.File{Name: "x.bin", Body: []byte("one")},
		tgtest.File{Name: "y.bin", Body: []byte("two")},
	)

	if !p.HasText("multi one") || !p.HasText("multi two") {
		t.Error("expect both files read back")
	}
}

func TestForm(t *testing.T) {
	app := newApp("index", func(p *tgframe.Params) error {
		form := tgcomp.Form(p.Main, &tgcomp.FormConf{Base: tgframe.Base{ID: "f"}})
		a := tgcomp.Textbox(form, "A")
		b := tgcomp.Checkbox(form, "B")
		tgcomp.Text(p.Main, fmt.Sprintf("a=%s b=%v", a, b))

		form2 := tgcomp.Form(p.Main, &tgcomp.FormConf{
			Base: tgframe.Base{ID: "g"}, HideSubmit: true})
		c := tgcomp.Textbox(form2, "C")
		if tgcomp.Button(form2, "Go") {
			tgcomp.Text(p.Main, "go c="+c)
		}
		return nil
	})

	p := tgtest.Open(t, app, "index")

	p.GetByLabel("A").Input("x")
	p.GetByLabel("B").Input(true)
	if !p.HasText("a= b=false") {
		t.Error("form values applied before submit")
	}

	forms := p.FindByName("form_component")
	if len(forms) != 2 {
		t.Fatalf("got %d forms, want 2", len(forms))
	}

	forms[0].Submit()
	if !p.HasText("a=x b=true") {
		t.Error("expect form values after submit")
	}

	// A button inside a form sends it.
	p.GetByLabel("C").Input("y")
	p.GetByLabel("Go").Click()
	if !p.HasText("go c=y") {
		t.Error("expect the click and value on one run")
	}
}

func TestErr(t *testing.T) {
	app := newApp("index", func(p *tgframe.Params) error {
		if tgcomp.Button(p.Main, "Fail") {
			return errors.New("boom")
		}
		return nil
	})

	p := tgtest.Open(t, app, "index")
	if err := p.Err(); err != nil {
		t.Fatalf("Err = %v", err)
	}

	p.GetByLabel("Fail").Click()
	if err := p.Err(); err == nil || !strings.Contains(err.Error(), "boom") {
		t.Errorf("Err = %v, want boom", err)
	}

	// A failed run leaves the last tree in place.
	if p.GetByLabel("Fail") == nil {
		t.Error("expect the button kept")
	}
}

func TestTreeDropsWhatRunDidNotDraw(t *testing.T) {
	app := newApp("index", func(p *tgframe.Params) error {
		show := tgcomp.Checkbox(p.Main, "Show")
		if show {
			box := tgcomp.Box(p.Main)
			tgcomp.Text(box, "inside")
		}
		return nil
	})

	p := tgtest.Open(t, app, "index")

	p.GetByLabel("Show").Input(true)
	if !p.HasText("inside") {
		t.Fatal("expect inside drawn")
	}

	// A held node is a snapshot, so the next run leaves it alone.
	held := p.Main()

	p.GetByLabel("Show").Input(false)
	if p.HasText("inside") {
		t.Error("expect inside gone")
	}

	if len(held.Children) != 2 {
		t.Errorf("held snapshot has %d children, want 2", len(held.Children))
	}
}

func TestMenuAndRadio(t *testing.T) {
	app := newApp("index", func(p *tgframe.Params) error {
		if i := tgcomp.Menu(p.Main, "Act", []string{"open", "save"}); i != nil {
			tgcomp.Text(p.Main, fmt.Sprintf("menu %d", *i))
		}

		if i := tgcomp.Radio(p.Main, "Size", []string{"s", "m"}); i != nil {
			tgcomp.Text(p.Main, fmt.Sprintf("radio %d", *i))
		}
		return nil
	})

	p := tgtest.Open(t, app, "index")

	p.GetByLabel("Act").Select(1)
	if !p.HasText("menu 1") {
		t.Error("expect menu 1")
	}

	p.GetByLabel("Size").Select(1)
	if !p.HasText("radio 1") {
		t.Error("expect radio 1")
	}
}

// A named form that moves keeps what it held, as React keeps it by id.
func TestFormQueueFollowsNamedForm(t *testing.T) {
	app := newApp("index", func(p *tgframe.Params) error {
		if !tgcomp.Checkbox(p.Sidebar, "Hide") {
			tgcomp.Text(p.Main, "banner")
		}

		form := tgcomp.Form(p.Main, &tgcomp.FormConf{Base: tgframe.Base{ID: "f"}})
		a := tgcomp.Textbox(form, "A")
		tgcomp.Text(p.Main, "a="+a)
		return nil
	})

	p := tgtest.Open(t, app, "index")

	p.GetByLabel("A").Input("x")
	p.GetByLabel("Hide").Input(true)

	p.FindByName("form_component")[0].Submit()
	if !p.HasText("a=x") {
		t.Error("expect the held value kept across the move")
	}
}

// A form gone from the page drops what it held.
func TestFormQueueDroppedWithForm(t *testing.T) {
	app := newApp("index", func(p *tgframe.Params) error {
		if !tgcomp.Checkbox(p.Sidebar, "Hide") {
			form := tgcomp.Form(p.Main, &tgcomp.FormConf{Base: tgframe.Base{ID: "f"}})
			a := tgcomp.Textbox(form, "A")
			tgcomp.Text(p.Main, "a="+a)
		}
		return nil
	})

	p := tgtest.Open(t, app, "index")

	p.GetByLabel("A").Input("x")
	p.GetByLabel("Hide").Input(true)
	p.GetByLabel("Hide").Input(false)

	p.FindByName("form_component")[0].Submit()
	if p.HasText("a=x") {
		t.Error("expect the held value dropped with the form")
	}
}

// A named node that moves on a failed run is not left in two places.
func TestMovedNodeOnFailedRun(t *testing.T) {
	app := newApp("index", func(p *tgframe.Params) error {
		hide := tgcomp.Checkbox(p.Sidebar, "Hide")
		if !hide {
			tgcomp.Text(p.Main, "banner")
		}

		tgcomp.Textbox(p.Main, "Name")
		if hide {
			return errors.New("fail")
		}
		return nil
	})

	p := tgtest.Open(t, app, "index")

	p.GetByLabel("Hide").Input(true)
	if p.Err() == nil {
		t.Fatal("expect the run to fail")
	}

	if got := len(p.Find(func(n *tgtest.Node) bool {
		return n.String("label") == "Name"
	})); got != 1 {
		t.Errorf("got %d Name textboxes, want 1", got)
	}
}

func TestHasTextOnlyDrawnText(t *testing.T) {
	app := newApp("index", func(p *tgframe.Params) error {
		tgcomp.FileUpload(p.Main, "Data", ".csv")
		tgcomp.Link(p.Main, "Docs", "https://example.com/docs")
		tgcomp.Menu(p.Main, "Act", []string{"open"})
		return nil
	})

	p := tgtest.Open(t, app, "index")

	for _, want := range []string{"Data", "Docs", "open"} {
		if !p.HasText(want) {
			t.Errorf("expect %q", want)
		}
	}

	item := p.GetByLabel("Act").Prop("items").([]any)[0].(map[string]any)
	for _, absent := range []string{".csv", "example.com", item["id"].(string)} {
		if p.HasText(absent) {
			t.Errorf("expect %q, which is never drawn as text, not to match", absent)
		}
	}
}

func TestSelectClear(t *testing.T) {
	app := newApp("index", func(p *tgframe.Params) error {
		i := tgcomp.Select(p.Main, "Fruit", []string{"apple", "pear"})
		tgcomp.Text(p.Main, fmt.Sprintf("picked %v", i != nil))
		return nil
	})

	p := tgtest.Open(t, app, "index")

	p.GetByLabel("Fruit").Select(0)
	p.GetByLabel("Fruit").Select(-1)
	if !p.HasText("picked false") {
		t.Error("expect Select(-1) to clear the select")
	}
}

func TestHasTextShownValue(t *testing.T) {
	app := newApp("index", func(p *tgframe.Params) error {
		tgcomp.Textbox(p.Main, "Name", &tgcomp.TextboxConf{Default: "Alice"})
		tgcomp.Textarea(p.Main, "Note", &tgcomp.TextareaConf{Default: "draft"})
		tgcomp.CodeInput(p.Main, "Script", &tgcomp.CodeInputConf{Default: "fmt.Println()"})
		tgcomp.Textbox(p.Main, "Pass", &tgcomp.TextboxConf{Default: "secret", Password: true})
		return nil
	})

	p := tgtest.Open(t, app, "index")

	if !p.HasText("Alice") || !p.HasText("draft") || !p.HasText("fmt.Println()") {
		t.Error("expect the defaults shown")
	}
	if p.HasText("secret") {
		t.Error("expect a password not shown")
	}

	p.GetByLabel("Name").Input("Bob")
	if !p.HasText("Bob") || p.HasText("Alice") {
		t.Error("expect the typed value shown in place of the default")
	}

	p.GetByLabel("Script").Input("")
	if p.HasText("fmt.Println()") {
		t.Error("expect a cleared code input shown empty")
	}
}
