package tgframe_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/voilelab/toolgui/toolgui/tgcomp"
	"github.com/voilelab/toolgui/toolgui/tgframe"
)

// pack is a notify pack as the GUI client reads it, cut down to what a test
// asserts on.
type pack struct {
	Type      int    `json:"type"`
	Key       string `json:"key"`
	Component struct {
		Name string `json:"name"`
		ID   string `json:"id"`
	} `json:"component"`
}

// String is what a failing test prints: the pack as "<what> <key>".
func (p pack) String() string {
	what := p.Component.Name
	switch p.Type {
	case tgframe.NotifyTypeDelete:
		what = "delete"
	case tgframe.NotifyTypeUpdate:
		what = "update " + what
	}

	return fmt.Sprintf("%s %s", what, p.Key)
}

// packsOf runs one page on state and returns every pack it sent, in order.
func packsOf(t *testing.T, state *tgframe.State, page tgframe.RunFunc) ([]pack, error) {
	t.Helper()

	app := tgframe.NewApp()
	app.AddPage("index", "Index", page)

	var packs []pack
	err := app.RunWithHandlingPanic("index", state, func(p tgframe.NotifyPack) {
		bs, mErr := json.Marshal(p)
		if mErr != nil {
			t.Fatalf("marshal pack: %v", mErr)
		}

		var one pack
		if uErr := json.Unmarshal(bs, &one); uErr != nil {
			t.Fatalf("unmarshal pack: %v", uErr)
		}

		packs = append(packs, one)
	})

	return packs, err
}

func wantPacks(t *testing.T, got []pack, want []string) {
	t.Helper()

	if len(got) != len(want) {
		t.Fatalf("got %d packs, want %d:\n got %v\nwant %v", len(got), len(want), got, want)
	}

	for i := range want {
		if got[i].String() != want[i] {
			t.Errorf("pack %d = %q, want %q", i, got[i].String(), want[i])
		}
	}
}

func TestEmptyRewritesRatherThanAppends(t *testing.T) {
	got, err := packsOf(t, tgframe.NewState(), func(p *tgframe.Params) error {
		slot := tgcomp.Empty(p.Main)
		slot.With(func(c *tgframe.Container) {
			tgcomp.Text(c, "Querying…")
		})
		slot.With(func(c *tgframe.Container) {
			tgcomp.Text(c, "Done")
		})
		return nil
	})
	if err != nil {
		t.Fatalf("run: %v", err)
	}

	// The slot's own key is written, cleared and written again: what the first
	// With put there is off the screen before the second one draws.
	wantPacks(t, got, []string{
		"empty_component container_component_container_main/0",
		// Whatever an earlier run left in the slot, this run starts empty.
		"delete container_component_container_main/0/0",
		"container_component container_component_container_main/0/0",
		"text_component container_component_container_main/0/0/0",
		"delete container_component_container_main/0/0",
		"container_component container_component_container_main/0/0",
		"text_component container_component_container_main/0/0/0",
	})
}

func TestEmptyClearLeavesTheSlotEmpty(t *testing.T) {
	got, err := packsOf(t, tgframe.NewState(), func(p *tgframe.Params) error {
		slot := tgcomp.Empty(p.Main)
		slot.With(func(c *tgframe.Container) {
			tgcomp.Text(c, "Querying…")
		})
		slot.Clear()

		// Clearing twice sends nothing: the slot is already empty.
		slot.Clear()
		return nil
	})
	if err != nil {
		t.Fatalf("run: %v", err)
	}

	wantPacks(t, got, []string{
		"empty_component container_component_container_main/0",
		"delete container_component_container_main/0/0",
		"container_component container_component_container_main/0/0",
		"text_component container_component_container_main/0/0/0",
		"delete container_component_container_main/0/0",
	})
}

func TestEmptyGivesBackTheIDsItClears(t *testing.T) {
	// The same widget twice in a row is a duplicated id anywhere else. In a
	// slot the first one is gone by the time the second is written, so it is
	// the same widget rather than two of them.
	_, err := packsOf(t, tgframe.NewState(), func(p *tgframe.Params) error {
		slot := tgcomp.Empty(p.Main)
		for range 3 {
			slot.With(func(c *tgframe.Container) {
				tgcomp.Textbox(c, "Name")
			})
		}
		return nil
	})
	if err != nil {
		t.Fatalf("run: %v", err)
	}

	// A widget that outlives the slot still collides with the one inside it:
	// giving an id back is not the same as never claiming it.
	_, err = packsOf(t, tgframe.NewState(), func(p *tgframe.Params) error {
		slot := tgcomp.Empty(p.Main)
		slot.With(func(c *tgframe.Container) {
			tgcomp.Textbox(c, "Name")
		})
		tgcomp.Textbox(p.Main, "Name")
		return nil
	})
	if !errors.Is(err, tgframe.ErrDuplicatedID) {
		t.Fatalf("err = %v, want ErrDuplicatedID", err)
	}
}

func TestClearedWidgetLeavesNoGhostState(t *testing.T) {
	const key = "textbox_component_Name"

	state := tgframe.NewState()
	state.Set(key, "typed by the user")

	// The widget is on the screen at the end of the run, so its value is the
	// one the user typed.
	_, err := packsOf(t, state, func(p *tgframe.Params) error {
		slot := tgcomp.Empty(p.Main)
		slot.With(func(c *tgframe.Container) {
			if got := tgcomp.Textbox(c, "Name"); got != "typed by the user" {
				t.Errorf("textbox = %q, want the value the user typed", got)
			}
		})
		return nil
	})
	if err != nil {
		t.Fatalf("run: %v", err)
	}

	if got, ok := state.Get[string](key); !ok || got != "typed by the user" {
		t.Fatalf("state[%s] = %q, %v; a widget still on the screen keeps its value", key, got, ok)
	}

	// Cleared off the screen, it does not hand that value to the next run.
	_, err = packsOf(t, state, func(p *tgframe.Params) error {
		slot := tgcomp.Empty(p.Main)
		slot.With(func(c *tgframe.Container) {
			tgcomp.Textbox(c, "Name")
		})
		slot.Clear()
		return nil
	})
	if err != nil {
		t.Fatalf("run: %v", err)
	}

	if _, ok := state.Get[string](key); ok {
		t.Fatalf("state[%s] is still set; a cleared widget should release its key", key)
	}
}

func TestSpinnerComesDownOnPanic(t *testing.T) {
	got, err := packsOf(t, tgframe.NewState(), func(p *tgframe.Params) error {
		defer tgcomp.Spinner(p.Main, "Loading…")()
		panic("page function blew up")
	})
	if !errors.Is(err, tgframe.ErrPanic) {
		t.Fatalf("err = %v, want ErrPanic", err)
	}

	wantPacks(t, got, []string{
		"empty_component container_component_container_main/0",
		"delete container_component_container_main/0/0",
		"container_component container_component_container_main/0/0",
		"spinner_component container_component_container_main/0/0/0",
		// The deferred call runs while the panic unwinds, so the spinner is
		// gone before the page reports the failure.
		"delete container_component_container_main/0/0",
	})
}

func TestStatusRewritesItselfOnEveryChange(t *testing.T) {
	got, err := packsOf(t, tgframe.NewState(), func(p *tgframe.Params) error {
		s := tgcomp.Status(p.Main, "Importing…")
		s.Write("one.csv")
		s.Complete("Imported")
		return nil
	})
	if err != nil {
		t.Fatalf("run: %v", err)
	}

	// Three renders — the first draw, the line, the close — each over the last.
	// The expander keeps one id through all of them, so the client keeps it
	// open across the change of label.
	wantPacks(t, got, []string{
		"empty_component container_component_container_main/0",
		"delete container_component_container_main/0/0",

		"container_component container_component_container_main/0/0",
		"expand_component container_component_container_main/0/0/0",
		"container_component container_component_container_main/0/0/0/0",

		"delete container_component_container_main/0/0",
		"container_component container_component_container_main/0/0",
		"expand_component container_component_container_main/0/0/0",
		"container_component container_component_container_main/0/0/0/0",
		"text_component container_component_container_main/0/0/0/0/0",

		"delete container_component_container_main/0/0",
		"container_component container_component_container_main/0/0",
		"expand_component container_component_container_main/0/0/0",
		"container_component container_component_container_main/0/0/0/0",
		"text_component container_component_container_main/0/0/0/0/0",
	})

	for _, p := range got {
		if p.Component.Name == "expand_component" && p.Component.ID != "expand_component_Importing…" {
			t.Errorf("expander id = %q, want it unchanged by the status' state", p.Component.ID)
		}
	}
}
