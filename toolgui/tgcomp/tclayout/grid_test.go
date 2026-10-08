package tclayout_test

import (
	"testing"

	"github.com/voilelab/toolgui/toolgui/tgcomp"
	"github.com/voilelab/toolgui/toolgui/tgframe"
)

// gridPacks keeps the packs carrying the grid component itself.
func gridPacks(packs []pack) []pack {
	var got []pack
	for _, p := range packs {
		if p.Component["name"] == "grid_component" {
			got = append(got, p)
		}
	}
	return got
}

func TestGridDefaults(t *testing.T) {
	packs := runPage(t, tgframe.NewState(), func(p *tgframe.Params) error {
		tgcomp.Grid(p.Main)
		return nil
	})

	got := gridPacks(packs)
	if len(got) != 1 {
		t.Fatalf("got %d grid packs, want 1", len(got))
	}

	props := got[0].Component
	if props["min_col_width"] != float64(tgcomp.GridDefaultMinColWidth) {
		t.Errorf("min_col_width = %v, want %v",
			props["min_col_width"], tgcomp.GridDefaultMinColWidth)
	}

	if props["gap"] != tgcomp.GridGapMD {
		t.Errorf("gap = %v, want %v", props["gap"], tgcomp.GridGapMD)
	}
}

func TestGridConf(t *testing.T) {
	packs := runPage(t, tgframe.NewState(), func(p *tgframe.Params) error {
		g := tgcomp.Grid(p.Main, &tgcomp.GridConf{
			ID:          "kpi",
			MinColWidth: 120,
			Gap:         tgcomp.GridGapXS,
		})
		tgcomp.Metric(g, "Revenue", "12.4M")
		return nil
	})

	props := gridPacks(packs)[0].Component
	if props["id"] != "grid_component_kpi" {
		t.Errorf("id = %v, want grid_component_kpi", props["id"])
	}

	if props["min_col_width"] != float64(120) {
		t.Errorf("min_col_width = %v, want 120", props["min_col_width"])
	}

	if props["gap"] != tgcomp.GridGapXS {
		t.Errorf("gap = %v, want %v", props["gap"], tgcomp.GridGapXS)
	}

	if !hasComponent(packs, "metric_component") {
		t.Error("what was written into the grid was not sent")
	}
}

func TestGridUnsupportedGapPanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("a gap the client cannot render did not panic")
		}
	}()

	c := tgframe.NewContainer("main", tgframe.NewState(), func(tgframe.NotifyPack) {})
	tgcomp.Grid(c, &tgcomp.GridConf{Gap: "huge"})
}

// An input in a grid in a form still reads the value the form submitted.
func TestGridInAForm(t *testing.T) {
	state := tgframe.NewState()
	page := func(got *string) tgframe.RunFunc {
		return func(p *tgframe.Params) error {
			f := tgcomp.Form(p.Main, &tgcomp.FormConf{ID: "search"})
			g := tgcomp.Grid(f, &tgcomp.GridConf{ID: "fields"})
			*got = tgcomp.Textbox(g, "Keyword")
			return nil
		}
	}

	var got string
	runPage(t, state, page(&got))

	event := &tgframe.EventForm{Events: []tgframe.Event{
		&tgframe.EventInput{ID: "textbox_component_Keyword", Value: "toolgui"},
	}}
	event.ApplyState(state)

	runPage(t, state, page(&got))
	if got != "toolgui" {
		t.Errorf("textbox in the grid = %q, want %q", got, "toolgui")
	}
}

// A grid in a slot is the slot's to take back, so the same grid may be
// written again without a duplicated id.
func TestGridInASlotIsCleared(t *testing.T) {
	app := tgframe.NewApp()
	app.AddPage("index", "Index", func(p *tgframe.Params) error {
		slot := tgcomp.Empty(p.Main)

		write := func(c *tgframe.Container) {
			g := tgcomp.Grid(c, &tgcomp.GridConf{ID: "kpi"})
			tgcomp.Metric(g, "Revenue", "12.4M", &tgcomp.MetricConf{ID: "revenue"})
		}

		slot.With(write)
		slot.With(write)
		return nil
	})

	err := app.Run("index", tgframe.NewState(), func(tgframe.NotifyPack) {})
	if err != nil {
		t.Fatalf("rewriting a slot holding a grid: %v", err)
	}
}

func TestGridInADialog(t *testing.T) {
	packs := runPage(t, tgframe.NewState(), func(p *tgframe.Params) error {
		d := tgcomp.Dialog(p.Main, "Delete")
		d.Open()
		d.With(func(c *tgframe.Container) {
			g := tgcomp.Grid(c)
			tgcomp.Metric(g, "Rows", "12")
		})
		return nil
	})

	if !hasComponent(packs, "grid_component") ||
		!hasComponent(packs, "metric_component") {
		t.Error("a grid in an open dialog was not written")
	}
}
