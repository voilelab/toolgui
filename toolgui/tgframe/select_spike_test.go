package tgframe_test

import (
	"errors"
	"testing"

	"github.com/voilelab/toolgui/toolgui/tgcomp"
	"github.com/voilelab/toolgui/toolgui/tgframe"
)

// TestTwoSelectsWithTheSameLabelNeedAnID is the case TG-12 was opened for.
// Select derives its id from its label and used to have no variant that took
// one, so a page with two same-label selects failed with no way out. Now the
// conf carries an ID, and giving one to either of them is the way out.
func TestTwoSelectsWithTheSameLabelNeedAnID(t *testing.T) {
	items := []string{"a", "b"}

	t.Run("neither has an id", func(t *testing.T) {
		_, _, err := keysOf(t, func(p *tgframe.Params) error {
			tgcomp.Select(p.Main, "Pick", items)
			tgcomp.Select(p.Main, "Pick", items)
			return nil
		})

		if !errors.Is(err, tgframe.ErrDuplicatedID) {
			t.Fatalf("err = %v, want ErrDuplicatedID", err)
		}
	})

	t.Run("the second has an id", func(t *testing.T) {
		_, ids, err := keysOf(t, func(p *tgframe.Params) error {
			tgcomp.Select(p.Main, "Pick", items)
			tgcomp.Select(p.Main, "Pick", items, &tgcomp.SelectConf{ID: "second"})
			return nil
		})
		if err != nil {
			t.Fatalf("run: %v", err)
		}

		want := []string{"select_component_Pick", "select_component_second"}
		for i := range want {
			if ids[i] != want[i] {
				t.Errorf("id %d = %q, want %q", i, ids[i], want[i])
			}
		}
	})
}

// TestSelectReadsTheStateUnderItsOwnID checks the id from the conf is the key
// the value comes back under, not just a label on the node.
func TestSelectReadsTheStateUnderItsOwnID(t *testing.T) {
	state := tgframe.NewState()

	// The client sends a 1-based index, 0 meaning nothing picked.
	state.Set("select_component_second", 2)

	var got *int
	app := tgframe.NewApp()
	app.AddPage("index", "Index", func(p *tgframe.Params) error {
		items := []string{"a", "b"}
		tgcomp.Select(p.Main, "Pick", items)
		got = tgcomp.Select(p.Main, "Pick", items, &tgcomp.SelectConf{ID: "second"})
		return nil
	})

	if err := app.Run("index", state, func(tgframe.NotifyPack) {}); err != nil {
		t.Fatalf("run: %v", err)
	}

	if got == nil || *got != 1 {
		t.Fatalf("got %v, want 1", got)
	}
}
