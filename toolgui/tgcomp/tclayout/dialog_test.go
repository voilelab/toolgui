package tclayout_test

import (
	"testing"

	"github.com/voilelab/toolgui/toolgui/tgcomp"
	"github.com/voilelab/toolgui/toolgui/tgframe"
	"github.com/voilelab/toolgui/toolgui/tgjson"
)

// dialogID is the id Dialog derives from its title, which is also the state
// key the openness is kept under.
const dialogID = "dialog_component_Delete"

// pack is a notify pack as the GUI client reads it, cut down to what these
// tests assert on.
type pack struct {
	Type      int            `json:"type"`
	Key       string         `json:"key"`
	Component map[string]any `json:"component"`
}

// runPage runs one page on state and returns every pack it sent, in order.
func runPage(t *testing.T, state *tgframe.State, page tgframe.RunFunc) []pack {
	t.Helper()

	app := tgframe.NewApp()
	app.AddPage("index", "Index", page)

	var packs []pack
	err := app.Run("index", state, func(p tgframe.NotifyPack) {
		bs, mErr := tgjson.Marshal(p)
		if mErr != nil {
			t.Fatalf("marshal pack: %v", mErr)
		}

		var one pack
		if uErr := tgjson.Unmarshal(bs, &one); uErr != nil {
			t.Fatalf("unmarshal pack: %v", uErr)
		}

		packs = append(packs, one)
	})
	if err != nil {
		t.Fatalf("run: %v", err)
	}

	return packs
}

// dialogPacks keeps the packs carrying the dialog component itself.
func dialogPacks(packs []pack) []pack {
	var got []pack
	for _, p := range packs {
		if p.Component["name"] == "dialog_component" {
			got = append(got, p)
		}
	}
	return got
}

func hasComponent(packs []pack, name string) bool {
	for _, p := range packs {
		if p.Component["name"] == name {
			return true
		}
	}
	return false
}

func TestDialogStartsClosedAndSkipsItsBody(t *testing.T) {
	packs := runPage(t, tgframe.NewState(), func(p *tgframe.Params) error {
		d := tgcomp.Dialog(p.Main, "Delete")
		d.With(func(c *tgframe.Container) {
			// A closed dialog never computes what it would have shown, so a
			// run that reaches here is the bug this guards against.
			panic("body of a closed dialog ran")
		})

		if d.IsOpen() {
			t.Error("IsOpen() = true on a dialog nothing opened")
		}
		return nil
	})

	// The node is sent anyway: Open() addresses it by key, so the key has to
	// exist even while nothing is on screen.
	got := dialogPacks(packs)
	if len(got) != 1 {
		t.Fatalf("got %d dialog packs, want 1", len(got))
	}

	if got[0].Type != tgframe.NotifyTypeCreate {
		t.Errorf("pack type = %d, want create", got[0].Type)
	}

	if got[0].Component["opened"] != false {
		t.Errorf("opened = %v, want false", got[0].Component["opened"])
	}
}

func TestDialogDefaults(t *testing.T) {
	packs := runPage(t, tgframe.NewState(), func(p *tgframe.Params) error {
		tgcomp.Dialog(p.Main, "Delete")
		return nil
	})

	props := dialogPacks(packs)[0].Component
	if props["id"] != dialogID {
		t.Errorf("id = %v, want %v", props["id"], dialogID)
	}

	if props["width"] != tgcomp.DialogWidthSmall {
		t.Errorf("width = %v, want %v", props["width"], tgcomp.DialogWidthSmall)
	}

	if props["dismissible"] != true {
		t.Errorf("dismissible = %v, want true", props["dismissible"])
	}
}

func TestDialogConf(t *testing.T) {
	packs := runPage(t, tgframe.NewState(), func(p *tgframe.Params) error {
		tgcomp.Dialog(p.Main, "Delete", (&tgcomp.DialogConf{
			ID:    "confirm",
			Width: tgcomp.DialogWidthLarge,
		}).SetDismissible(false))
		return nil
	})

	props := dialogPacks(packs)[0].Component
	if props["id"] != "dialog_component_confirm" {
		t.Errorf("id = %v, want dialog_component_confirm", props["id"])
	}

	if props["width"] != tgcomp.DialogWidthLarge {
		t.Errorf("width = %v, want %v", props["width"], tgcomp.DialogWidthLarge)
	}

	if props["dismissible"] != false {
		t.Errorf("dismissible = %v, want false", props["dismissible"])
	}
}

func TestDialogUnsupportedWidthPanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("a width the client cannot render did not panic")
		}
	}()

	c := tgframe.NewContainer("main", tgframe.NewState(), func(tgframe.NotifyPack) {})
	tgcomp.Dialog(c, "Delete", &tgcomp.DialogConf{Width: "huge"})
}

func TestDialogOpenBeforeWithDrawsTheBody(t *testing.T) {
	packs := runPage(t, tgframe.NewState(), func(p *tgframe.Params) error {
		d := tgcomp.Dialog(p.Main, "Delete")
		d.Open()

		if !d.IsOpen() {
			t.Error("IsOpen() = false right after Open()")
		}

		d.With(func(c *tgframe.Container) {
			tgcomp.Text(c, "Delete it?")
		})
		return nil
	})

	got := dialogPacks(packs)
	if len(got) != 2 {
		t.Fatalf("got %d dialog packs, want 2 (create, update)", len(got))
	}

	if got[1].Type != tgframe.NotifyTypeUpdate {
		t.Errorf("second pack type = %d, want update", got[1].Type)
	}

	if got[1].Component["opened"] != true {
		t.Errorf("opened = %v, want true", got[1].Component["opened"])
	}

	if !hasComponent(packs, "text_component") {
		t.Error("the body of an opened dialog was not written")
	}
}

func TestDialogOpenAfterWithLeavesTheBodyForTheNextRun(t *testing.T) {
	state := tgframe.NewState()
	packs := runPage(t, state, func(p *tgframe.Params) error {
		d := tgcomp.Dialog(p.Main, "Delete")
		d.With(func(c *tgframe.Container) {
			tgcomp.Text(c, "Delete it?")
		})

		// Openness is a prop rather than "was the node sent", so this opens
		// the dialog on the client even though With has already gone by.
		d.Open()
		return nil
	})

	got := dialogPacks(packs)
	if len(got) != 2 || got[1].Component["opened"] != true {
		t.Fatalf("Open() after With() did not send an opened update: %v", got)
	}

	// The body belongs to this run's With, which ran while the dialog was
	// still closed. It arrives on the next one.
	if hasComponent(packs, "text_component") {
		t.Error("the body was written by a With that ran while closed")
	}

	next := runPage(t, state, func(p *tgframe.Params) error {
		d := tgcomp.Dialog(p.Main, "Delete")
		d.With(func(c *tgframe.Container) {
			tgcomp.Text(c, "Delete it?")
		})
		return nil
	})

	if !hasComponent(next, "text_component") {
		t.Error("the dialog opened last run did not draw its body on this one")
	}
}

func TestDialogOpennessSurvivesTheRun(t *testing.T) {
	state := tgframe.NewState()
	runPage(t, state, func(p *tgframe.Params) error {
		tgcomp.Dialog(p.Main, "Delete").Open()
		return nil
	})

	// Kept under the dialog's id as a bare bool, which is what the client
	// writes back through an input event.
	if opened, _ := state.Get[bool](dialogID); !opened {
		t.Fatal("Open() did not leave the dialog open in the state")
	}

	packs := runPage(t, state, func(p *tgframe.Params) error {
		d := tgcomp.Dialog(p.Main, "Delete")
		if !d.IsOpen() {
			t.Error("IsOpen() = false on the run after Open()")
		}
		return nil
	})

	if dialogPacks(packs)[0].Component["opened"] != true {
		t.Error("the dialog opened last run was sent closed")
	}
}

func TestDialogCloseStaysClosedOnTheNextRun(t *testing.T) {
	state := tgframe.NewState()
	runPage(t, state, func(p *tgframe.Params) error {
		tgcomp.Dialog(p.Main, "Delete").Open()
		return nil
	})

	packs := runPage(t, state, func(p *tgframe.Params) error {
		d := tgcomp.Dialog(p.Main, "Delete")
		d.With(func(c *tgframe.Container) {
			tgcomp.Text(c, "Delete it?")
			d.Close()
		})
		return nil
	})

	// The body was already sent when Close ran, so the client is told
	// separately; this run does not take it back.
	got := dialogPacks(packs)
	if len(got) != 2 || got[1].Component["opened"] != false {
		t.Fatalf("Close() did not send a closed update: %v", got)
	}

	if !hasComponent(packs, "text_component") {
		t.Error("the body drawn before Close() was not sent")
	}

	next := runPage(t, state, func(p *tgframe.Params) error {
		d := tgcomp.Dialog(p.Main, "Delete")
		d.With(func(c *tgframe.Container) {
			panic("body of a closed dialog ran")
		})
		return nil
	})

	if dialogPacks(next)[0].Component["opened"] != false {
		t.Error("a closed dialog came back open on the next run")
	}
}

func TestDialogManualCloseIsNotUndoneByTheNextRun(t *testing.T) {
	state := tgframe.NewState()
	runPage(t, state, func(p *tgframe.Params) error {
		tgcomp.Dialog(p.Main, "Delete").Open()
		return nil
	})

	// What the client sends when the app user dismisses the dialog.
	event := &tgframe.EventInput{ID: dialogID, Value: false}
	event.ApplyState(state)

	packs := runPage(t, state, func(p *tgframe.Params) error {
		d := tgcomp.Dialog(p.Main, "Delete")
		d.With(func(c *tgframe.Container) {
			panic("body of a closed dialog ran")
		})
		return nil
	})

	if dialogPacks(packs)[0].Component["opened"] != false {
		t.Error("a dialog the app user closed was reopened by the next run")
	}
}

func TestDialogBodyKeepsItsWidgetStateWhileClosed(t *testing.T) {
	state := tgframe.NewState()
	runPage(t, state, func(p *tgframe.Params) error {
		d := tgcomp.Dialog(p.Main, "Delete")
		d.Open()
		d.With(func(c *tgframe.Container) {
			tgcomp.Textbox(c, "Reason")
		})
		return nil
	})

	event := &tgframe.EventInput{ID: "textbox_component_Reason", Value: "typo"}
	event.ApplyState(state)

	// Closed, so the body is not written: its widgets are hidden rather than
	// gone, the same as anything behind an `if`.
	runPage(t, state, func(p *tgframe.Params) error {
		d := tgcomp.Dialog(p.Main, "Delete")
		d.Close()
		d.With(func(c *tgframe.Container) {
			tgcomp.Textbox(c, "Reason")
		})
		return nil
	})

	var typed string
	runPage(t, state, func(p *tgframe.Params) error {
		d := tgcomp.Dialog(p.Main, "Delete")
		d.Open()
		d.With(func(c *tgframe.Container) {
			typed = tgcomp.Textbox(c, "Reason")
		})
		return nil
	})

	if typed != "typo" {
		t.Errorf("textbox in the dialog = %q, want %q", typed, "typo")
	}
}
