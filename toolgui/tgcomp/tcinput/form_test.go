package tcinput_test

import (
	"testing"

	"github.com/voilelab/toolgui/toolgui/tgcomp/tcinput"
	"github.com/voilelab/toolgui/toolgui/tgframe"
)

// formRun is what one run of the search form below saw.
type formRun struct {
	keyword  string
	searched bool
}

// newFormRunner is a form with no submit button of its own: a textbox and a
// button, the shape a page takes when the form's own button is the one that
// sends it.
func newFormRunner(t *testing.T, seen *formRun) *runner {
	t.Helper()

	return newRunner(t, func(p *tgframe.Params, _ *clickRun) {
		tcinput.Form(p.Main, &tcinput.FormConf{ID: "search", HideSubmit: true}).
			With(func(c *tgframe.Container) {
				seen.keyword = tcinput.Textbox(c, "keyword")
				seen.searched = tcinput.Button(c, "Search")
			})
	})
}

// TestFormButtonSubmits is what a button inside a form is for: the click
// arrives with the inputs that were made before it, so the run that reports
// the click is also the one that reads the new values.
func TestFormButtonSubmits(t *testing.T) {
	var seen formRun
	r := newFormRunner(t, &seen)

	r.run(&tgframe.EventForm{Events: []tgframe.Event{
		&tgframe.EventInput{ID: "textbox_component_keyword", Value: "toolgui"},
		&tgframe.EventClick{ID: "button_component_Search"},
	}})

	if !seen.searched {
		t.Error("Button = false on the run its click was submitted with," +
			" want true")
	}

	if seen.keyword != "toolgui" {
		t.Errorf("Textbox = %q, want %q", seen.keyword, "toolgui")
	}
}

// TestFormKeepsValuesAfterSubmit pins the half that is not the click: the
// values stay in the state, so the next run still reads them, and the click
// is gone.
func TestFormKeepsValuesAfterSubmit(t *testing.T) {
	var seen formRun
	r := newFormRunner(t, &seen)

	r.run(&tgframe.EventForm{Events: []tgframe.Event{
		&tgframe.EventInput{ID: "textbox_component_keyword", Value: "toolgui"},
		&tgframe.EventClick{ID: "button_component_Search"},
	}})

	r.run(&tgframe.EventEmpty{})

	if seen.searched {
		t.Error("Button = true on the run after the click, want false")
	}

	if seen.keyword != "toolgui" {
		t.Errorf("Textbox = %q on the run after the click, want %q",
			seen.keyword, "toolgui")
	}
}
