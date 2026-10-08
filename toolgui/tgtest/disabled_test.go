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

	if !expectFatal(t, p, func() { node.UploadFiles() }) {
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

func TestSelectDataFrameOutOfRange(t *testing.T) {
	app := tgframe.NewApp()
	app.AddPage("index", "index", func(p *tgframe.Params) error {
		tgcomp.DataFrame(p.Main, []string{"single"}, [][]string{{"a"}},
			&tgcomp.DataFrameConf{Selection: tgcomp.SelectionModeSingle})
		tgcomp.DataFrame(p.Main, []string{"multi"}, [][]string{{"a"}},
			&tgcomp.DataFrameConf{
				Selection: tgcomp.SelectionModeMulti,
				RowKeys:   []string{"k"},
			})
		return nil
	})

	p := Open(t, app, "index")
	frames := p.FindByName("dataframe_component")
	single, multi := frames[0], frames[1]

	cases := map[string]func(){
		"index past the rows":       func() { single.Select(1) },
		"negative index":            func() { single.Select(-1) },
		"keyed index past the rows": func() { multi.SelectMany(0, 1) },
		"Select on multi":           func() { multi.Select(0) },
		"SelectMany on single":      func() { single.SelectMany(0) },
	}
	for name, f := range cases {
		if !expectFatal(t, p, f) {
			t.Errorf("%s: expect a fail", name)
		}
	}
}
