package tcinput_test

import (
	"testing"

	"github.com/voilelab/toolgui/toolgui/tgcomp/tcinput"
	"github.com/voilelab/toolgui/toolgui/tgframe"
)

// clickRun is what one run of a page below saw: what the Clicked getter said
// before the button was drawn, and what the button returned when it was.
type clickRun struct {
	before bool
	drawn  bool
}

// runner drives a page through a real session, so the click id under test is
// the one an event actually lands in the state.
type runner struct {
	t       *testing.T
	session *tgframe.Session
	done    chan *tgframe.ResultPack
	last    clickRun
}

// newRunner starts a session on page and renders it once, which is what a
// client does before it can click anything. page records what it saw in seen.
func newRunner(t *testing.T, page func(p *tgframe.Params, seen *clickRun)) *runner {
	t.Helper()

	r := &runner{t: t, done: make(chan *tgframe.ResultPack, 1)}

	app := tgframe.NewApp()
	app.AddPage("test", "Test", func(p *tgframe.Params) error {
		r.last = clickRun{}
		page(p, &r.last)
		return nil
	})

	session, err := tgframe.NewSession(app, "test", tgframe.NewState(),
		func(pack any) error {
			if result, ok := pack.(*tgframe.ResultPack); ok {
				r.done <- result
			}
			return nil
		})
	if err != nil {
		t.Fatalf("NewSession: %v", err)
	}

	r.session = session
	t.Cleanup(session.Close)

	r.run(&tgframe.EventEmpty{})

	return r
}

// run handles the event and returns what the page function saw.
func (r *runner) run(event tgframe.Event) clickRun {
	r.t.Helper()

	r.session.HandleEvent(event)

	result := <-r.done
	if !result.Success {
		r.t.Fatalf("run failed: %q", result.Error)
	}

	return r.last
}

// newButtonRunner is a page that asks about the button under id, then draws
// it — both from the same label and conf, which is the contract.
func newButtonRunner(t *testing.T, id string) *runner {
	t.Helper()

	return newRunner(t, func(p *tgframe.Params, seen *clickRun) {
		conf := &tcinput.ButtonConf{ID: id}
		seen.before = tcinput.ButtonClicked(p.Main, "Save", conf)
		seen.drawn = tcinput.Button(p.Main, "Save", conf)
	})
}

// TestButtonClickedBeforeDraw is the point of ButtonClicked: the page knows
// about the click while it can still decide what to draw.
func TestButtonClickedBeforeDraw(t *testing.T) {
	r := newButtonRunner(t, "save")

	got := r.run(&tgframe.EventClick{ID: "button_component_save"})
	if !got.before {
		t.Error("ButtonClicked = false before the button was drawn, want true")
	}

	if got.before != got.drawn {
		t.Errorf("ButtonClicked = %v, Button = %v, want the same",
			got.before, got.drawn)
	}
}

func TestButtonClickedFalse(t *testing.T) {
	// A conf id compared as-is never matches: the component id carries the
	// component name in front of it.
	t.Run("bare conf id on the event", func(t *testing.T) {
		r := newButtonRunner(t, "save")

		got := r.run(&tgframe.EventClick{ID: "save"})
		if got.before || got.drawn {
			t.Errorf("ButtonClicked = %v, Button = %v, want both false",
				got.before, got.drawn)
		}
	})

	t.Run("another button", func(t *testing.T) {
		r := newButtonRunner(t, "save")

		got := r.run(&tgframe.EventClick{ID: "button_component_cancel"})
		if got.before || got.drawn {
			t.Errorf("ButtonClicked = %v, Button = %v, want both false",
				got.before, got.drawn)
		}
	})

	// The click is this run's, not the state's: the next run clears it.
	t.Run("next run", func(t *testing.T) {
		r := newButtonRunner(t, "save")

		if got := r.run(&tgframe.EventClick{ID: "button_component_save"}); !got.before {
			t.Fatal("ButtonClicked = false on the click run, want true")
		}

		got := r.run(&tgframe.EventEmpty{})
		if got.before || got.drawn {
			t.Errorf("ButtonClicked = %v, Button = %v, want both false",
				got.before, got.drawn)
		}
	})
}

// TestButtonClickedNeverDrawn is what the id check is for. Button reports a
// click only where the button is written, so a made-up click id can only name
// something on the screen; ButtonClicked is asked before anything is written,
// so it has to check that itself — or a click id the client invented would run
// an action the page never offered.
func TestButtonClickedNeverDrawn(t *testing.T) {
	r := newRunner(t, func(p *tgframe.Params, seen *clickRun) {
		seen.before = tcinput.ButtonClicked(p.Main, "Delete",
			&tcinput.ButtonConf{ID: "delete"})
	})

	got := r.run(&tgframe.EventClick{ID: "button_component_delete"})
	if got.before {
		t.Error("ButtonClicked = true for a button the page never drew," +
			" want false")
	}
}

// TestDownloadButtonClicked pins the same behaviour for the download button,
// whose ids come from its own component name.
func TestDownloadButtonClicked(t *testing.T) {
	newDownloadRunner := func(t *testing.T) *runner {
		t.Helper()

		return newRunner(t, func(p *tgframe.Params, seen *clickRun) {
			conf := &tcinput.DownloadButtonConf{ID: "report"}
			seen.before = tcinput.DownloadButtonClicked(p.Main, "Report", conf)
			seen.drawn = tcinput.DownloadButton(p.Main, "Report",
				[]byte("body"), conf)
		})
	}

	t.Run("clicked", func(t *testing.T) {
		r := newDownloadRunner(t)

		got := r.run(&tgframe.EventClick{ID: "download_button_component_report"})
		if !got.before {
			t.Error("DownloadButtonClicked = false before the draw, want true")
		}

		if got.before != got.drawn {
			t.Errorf("DownloadButtonClicked = %v, DownloadButton = %v,"+
				" want the same", got.before, got.drawn)
		}
	})

	t.Run("another button", func(t *testing.T) {
		r := newDownloadRunner(t)

		got := r.run(&tgframe.EventClick{ID: "download_button_component_other"})
		if got.before || got.drawn {
			t.Errorf("DownloadButtonClicked = %v, DownloadButton = %v,"+
				" want both false", got.before, got.drawn)
		}
	})

	t.Run("never drawn", func(t *testing.T) {
		r := newRunner(t, func(p *tgframe.Params, seen *clickRun) {
			seen.before = tcinput.DownloadButtonClicked(p.Main, "Secret",
				&tcinput.DownloadButtonConf{ID: "secret"})
		})

		got := r.run(&tgframe.EventClick{ID: "download_button_component_secret"})
		if got.before {
			t.Error("DownloadButtonClicked = true for a button the page never" +
				" drew, want false")
		}
	})
}

// TestClickedDrawsNothing is the other half of reading before the draw: the
// getters take the same arguments the draw does, and must not put a button on
// the screen for it.
func TestClickedDrawsNothing(t *testing.T) {
	var packs int
	c := tgframe.NewContainer("test", tgframe.NewState(),
		func(pack tgframe.NotifyPack) { packs++ })

	_ = tcinput.ButtonClicked(c, "Save", &tcinput.ButtonConf{ID: "save"})
	_ = tcinput.DownloadButtonClicked(c, "Report")
	_ = tcinput.DownloadFileClicked(c, "Archive")

	if packs != 0 {
		t.Errorf("drew %d components, want none", packs)
	}
}
