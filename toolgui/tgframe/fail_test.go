package tgframe_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/voilelab/toolgui/toolgui/tgcomp"
	"github.com/voilelab/toolgui/toolgui/tgframe"
)

// namesOf runs one page and returns the typename of every component it sent,
// in order, along with what App.Run made of the run.
func namesOf(t *testing.T, page tgframe.RunFunc) ([]string, error) {
	t.Helper()

	var pack struct {
		Type      int `json:"type"`
		Component struct {
			Name    string `json:"name"`
			Message string `json:"message"`
		} `json:"component"`
	}

	var names []string
	app := tgframe.NewApp()
	app.AddPage("index", "Index", page)

	err := app.Run("index", tgframe.NewState(), func(p tgframe.NotifyPack) {
		bs, mErr := json.Marshal(p)
		if mErr != nil {
			t.Fatalf("marshal pack: %v", mErr)
		}
		if uErr := json.Unmarshal(bs, &pack); uErr != nil {
			t.Fatalf("unmarshal pack: %v", uErr)
		}
		if pack.Type != tgframe.NotifyTypeCreate {
			return
		}

		names = append(names, pack.Component.Name)
	})

	return names, err
}

// TestFailedComponentDoesNotStopThePage is the point of the whole thing: a
// Table whose rows do not match its head leaves a placeholder where it would
// have been, everything after it still renders, and App.Run reports it.
func TestFailedComponentDoesNotStopThePage(t *testing.T) {
	names, err := namesOf(t, func(p *tgframe.Params) error {
		tgcomp.Text(p.Main, "before")
		tgcomp.Table(p.Main, []string{"a", "b"}, [][]string{{"1"}})
		tgcomp.Text(p.Main, "after")
		return nil
	})
	if err == nil {
		t.Fatal("Run returned no error, want the table's failure")
	}

	if errors.Is(err, tgframe.ErrPanic) {
		t.Errorf("err = %v, want a plain run error rather than a panic", err)
	}

	want := []string{
		"text_component",
		tgframe.ErrorComponentName,
		"text_component",
	}
	if len(names) != len(want) {
		t.Fatalf("components = %v, want %v", names, want)
	}
	for i := range want {
		if names[i] != want[i] {
			t.Fatalf("components = %v, want %v", names, want)
		}
	}
}

// TestFirstFailureIsTheOneReported pins that a second failing component does
// not overwrite the first one's error, and still renders its own placeholder.
func TestFirstFailureIsTheOneReported(t *testing.T) {
	first := errors.New("the first failure")
	second := errors.New("the second failure")

	names, err := namesOf(t, func(p *tgframe.Params) error {
		p.Main.Fail(first)
		p.Main.Fail(second)
		tgcomp.Text(p.Main, "after")
		return nil
	})
	if !errors.Is(err, first) {
		t.Errorf("err = %v, want %v", err, first)
	}

	if len(names) != 3 {
		t.Fatalf("components = %v, want two placeholders and a text", names)
	}
}

// TestFailIsNotWhatThePageReturns keeps the page function's own error ahead of
// a component's: the page said the run was over, and that reason wins.
func TestFailIsNotWhatThePageReturns(t *testing.T) {
	pageErr := errors.New("the page gave up")

	_, err := namesOf(t, func(p *tgframe.Params) error {
		p.Main.Fail(errors.New("a component failed"))
		return pageErr
	})
	if !errors.Is(err, pageErr) {
		t.Errorf("err = %v, want %v", err, pageErr)
	}
}

// TestFailIgnoresNil checks a component that reports no failure adds nothing.
func TestFailIgnoresNil(t *testing.T) {
	names, err := namesOf(t, func(p *tgframe.Params) error {
		p.Main.Fail(nil)
		return nil
	})
	if err != nil {
		t.Errorf("Run: %v", err)
	}

	if len(names) != 0 {
		t.Errorf("components = %v, want none", names)
	}
}

// TestDuplicatedIDStillWins keeps ErrDuplicatedID's own path: it is recorded
// the same way, and a component failing afterwards does not replace it.
func TestDuplicatedIDStillWins(t *testing.T) {
	_, err := namesOf(t, func(p *tgframe.Params) error {
		tgcomp.Text(p.Main, "one", &tgcomp.TextConf{ID: "dup"})
		tgcomp.Text(p.Main, "two", &tgcomp.TextConf{ID: "dup"})
		p.Main.Fail(errors.New("a later failure"))
		return nil
	})
	if !errors.Is(err, tgframe.ErrDuplicatedID) {
		t.Errorf("err = %v, want %v", err, tgframe.ErrDuplicatedID)
	}
}

// TestCallerMistakeStillPanics is the other side of the split: a call no data
// can make right still panics, and RunWithHandlingPanic still wraps it.
func TestCallerMistakeStillPanics(t *testing.T) {
	app := tgframe.NewApp()
	app.AddPage("index", "Index", func(p *tgframe.Params) error {
		tgcomp.Column(p.Main, 0)
		return nil
	})

	err := app.RunWithHandlingPanic("index", tgframe.NewState(),
		func(p tgframe.NotifyPack) {})
	if !errors.Is(err, tgframe.ErrPanic) {
		t.Errorf("err = %v, want ErrPanic", err)
	}
}
