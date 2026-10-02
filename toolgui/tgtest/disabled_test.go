package tgtest

import (
	"testing"

	"github.com/voilelab/toolgui/toolgui/tgcomp"
	"github.com/voilelab/toolgui/toolgui/tgframe"
)

// fatalTB records Fatalf instead of stopping the test.
type fatalTB struct {
	testing.TB
	failed bool
}

func (f *fatalTB) Fatalf(string, ...any) {
	f.failed = true
	panic(f)
}

func TestClickDisabled(t *testing.T) {
	app := tgframe.NewApp()
	app.AddPage("index", "index", func(p *tgframe.Params) error {
		tgcomp.Button(p.Main, "Go", &tgcomp.ButtonConf{Disabled: true})
		return nil
	})

	p := Open(t, app, "index")
	node := p.GetByLabel("Go")

	tb := &fatalTB{TB: t}
	p.t = tb

	func() {
		defer func() {
			if r := recover(); r != nil && r != tb {
				panic(r)
			}
		}()
		node.Click()
	}()

	if !tb.failed {
		t.Error("expect a click on a disabled button to fail")
	}
}

func TestUploadEmptyPick(t *testing.T) {
	app := tgframe.NewApp()
	app.AddPage("index", "index", func(p *tgframe.Params) error {
		tgcomp.MultiFileUpload(p.Main, "Files", "")
		return nil
	})

	p := Open(t, app, "index")
	node := p.GetByLabel("Files")

	tb := &fatalTB{TB: t}
	p.t = tb

	func() {
		defer func() { recover() }()
		node.UploadFiles()
	}()

	if !tb.failed {
		t.Error("expect an empty pick to fail")
	}
}

// expectFatal reports whether f fails the test of p.
func expectFatal(t *testing.T, p *Page, f func()) bool {
	tb := &fatalTB{TB: t}
	p.t = tb
	defer func() { p.t = t }()

	func() {
		defer func() {
			if r := recover(); r != nil && r != tb {
				panic(r)
			}
		}()
		f()
	}()

	return tb.failed
}

func TestSelectOutOfRange(t *testing.T) {
	app := tgframe.NewApp()
	app.AddPage("index", "index", func(p *tgframe.Params) error {
		items := []string{"a", "b"}
		tgcomp.Select(p.Main, "Select", items)
		tgcomp.Radio(p.Main, "Radio", items)
		tgcomp.SelectSlider(p.Main, "Slider", items)
		tgcomp.Menu(p.Main, "Menu", items)
		tgcomp.Textbox(p.Main, "Box")
		return nil
	})

	p := Open(t, app, "index")

	for _, c := range []struct {
		label string
		i     int
	}{
		{"Select", -2}, {"Select", 2},
		{"Radio", -1}, {"Radio", 2},
		{"Slider", -1}, {"Slider", 2},
		{"Menu", -1}, {"Menu", 2},
		{"Box", 0},
	} {
		node := p.GetByLabel(c.label)
		if !expectFatal(t, p, func() { node.Select(c.i) }) {
			t.Errorf("expect %s.Select(%d) to fail", c.label, c.i)
		}
	}
}
