package tclayout_test

import (
	"testing"

	"github.com/voilelab/toolgui/toolgui/tgcomp"
	"github.com/voilelab/toolgui/toolgui/tgframe"
)

// toolbarPacks keeps the packs carrying the toolbar component itself.
func toolbarPacks(packs []pack) []pack {
	var got []pack
	for _, p := range packs {
		if p.Component["name"] == "toolbar_component" {
			got = append(got, p)
		}
	}
	return got
}

func TestToolbarDefaults(t *testing.T) {
	packs := runPage(t, tgframe.NewState(), func(p *tgframe.Params) error {
		tgcomp.Toolbar(p.Main)
		return nil
	})

	got := toolbarPacks(packs)
	if len(got) != 1 {
		t.Fatalf("got %d toolbar packs, want 1", len(got))
	}

	props := got[0].Component
	if props["sticky"] != false {
		t.Errorf("sticky = %v, want false", props["sticky"])
	}

	if props["justify"] != tgcomp.ToolbarJustifyStart {
		t.Errorf("justify = %v, want %v",
			props["justify"], tgcomp.ToolbarJustifyStart)
	}
}

func TestToolbarConf(t *testing.T) {
	packs := runPage(t, tgframe.NewState(), func(p *tgframe.Params) error {
		bar := tgcomp.Toolbar(p.Main, &tgcomp.ToolbarConf{
			ID:      "head",
			Sticky:  true,
			Justify: tgcomp.ToolbarJustifyBetween,
		})
		tgcomp.Text(bar, "in the row")
		return nil
	})

	props := toolbarPacks(packs)[0].Component
	if props["id"] != "toolbar_component_head" {
		t.Errorf("id = %v, want toolbar_component_head", props["id"])
	}

	if props["sticky"] != true {
		t.Errorf("sticky = %v, want true", props["sticky"])
	}

	if props["justify"] != tgcomp.ToolbarJustifyBetween {
		t.Errorf("justify = %v, want %v",
			props["justify"], tgcomp.ToolbarJustifyBetween)
	}

	if !hasComponent(packs, "text_component") {
		t.Error("what was written into the toolbar was not sent")
	}
}

func TestToolbarUnsupportedJustifyPanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("a justify the client cannot render did not panic")
		}
	}()

	c := tgframe.NewContainer("main", tgframe.NewState(), func(tgframe.NotifyPack) {})
	tgcomp.Toolbar(c, &tgcomp.ToolbarConf{Justify: "middle"})
}

// A toolbar in a slot is the slot's to take back: clearing it gives the ids
// inside up, so the same row may be written again without the run failing
// with a duplicated id.
func TestToolbarInASlotIsCleared(t *testing.T) {
	app := tgframe.NewApp()
	app.AddPage("index", "Index", func(p *tgframe.Params) error {
		slot := tgcomp.Empty(p.Main)

		write := func(c *tgframe.Container) {
			bar := tgcomp.Toolbar(c, &tgcomp.ToolbarConf{ID: "head"})
			tgcomp.Button(bar, "Run", &tgcomp.ButtonConf{ID: "run"})
		}

		slot.With(write)
		slot.With(write)
		return nil
	})

	err := app.Run("index", tgframe.NewState(), func(tgframe.NotifyPack) {})
	if err != nil {
		t.Fatalf("rewriting a slot holding a toolbar: %v", err)
	}
}
