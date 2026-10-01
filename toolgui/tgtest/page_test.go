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
