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
