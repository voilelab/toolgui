package tcinput_test

import (
	"testing"

	"github.com/voilelab/toolgui/toolgui/tgcomp/tcinput"
	"github.com/voilelab/toolgui/toolgui/tgframe"
)

// clickRun is what one run of the page below saw: what ButtonClicked said
// before the button was drawn, and what Button returned when it was.
type clickRun struct {
	before bool
	drawn  bool
}

// buttonRunner drives a one-button page through a real session, so the click
// id under test is the one an event actually lands in the state.
type buttonRunner struct {
	t       *testing.T
	session *tgframe.Session
	done    chan *tgframe.ResultPack
	last    clickRun
}

func newButtonRunner(t *testing.T, id string) *buttonRunner {
	t.Helper()

	r := &buttonRunner{t: t, done: make(chan *tgframe.ResultPack, 1)}

	app := tgframe.NewApp()
	app.AddPage("test", "Test", func(p *tgframe.Params) error {
		r.last = clickRun{before: tcinput.ButtonClicked(p.State, id)}
		r.last.drawn = tcinput.Button(p.Main, "Save", &tcinput.ButtonConf{ID: id})
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

	return r
}

// run handles the event and returns what the page function saw.
func (r *buttonRunner) run(event tgframe.Event) clickRun {
	r.t.Helper()

	r.session.HandleEvent(event)

	result := <-r.done
	if !result.Success {
		r.t.Fatalf("run failed: %q", result.Error)
	}

	return r.last
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

// TestDownloadButtonClicked pins the same behaviour for the download button,
// whose ids come from its own component name.
func TestDownloadButtonClicked(t *testing.T) {
	state := tgframe.NewState()
	state.SetClickID("download_button_component_report")

	if !tcinput.DownloadButtonClicked(state, "report") {
		t.Error("DownloadButtonClicked = false, want true")
	}

	if tcinput.DownloadButtonClicked(state, "other") {
		t.Error("DownloadButtonClicked = true for another id, want false")
	}

	// The button's own return value agrees with it.
	if !tcinput.DownloadButton(defaultContainer(state), "Report", []byte("body"),
		&tcinput.DownloadButtonConf{ID: "report"}) {

		t.Error("DownloadButton = false, want true")
	}
}
