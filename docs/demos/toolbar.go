package demos

import (
	"fmt"

	"github.com/voilelab/toolgui/toolgui/tgcomp"
	"github.com/voilelab/toolgui/toolgui/tgframe"
)

func toolbarDemo(p *tgframe.Params) error {
	// ANCHOR: demo
	bar := tgcomp.Toolbar(p.Main, &tgcomp.ToolbarConf{ID: "toolbar"})

	run := tgcomp.Button(bar, "Run", &tgcomp.ButtonConf{ID: "toolbar_run"})
	tgcomp.Button(bar, "Stop", &tgcomp.ButtonConf{ID: "toolbar_stop"})
	modes := []string{"fast", "safe"}
	mode := tgcomp.Select(bar, "Mode", modes,
		(&tgcomp.SelectConf{ID: "toolbar_mode"}).SetDefault(0))

	tgcomp.Text(p.Main, fmt.Sprintf("Run: %v, mode: %s", run, modes[*mode]),
		&tgcomp.TextConf{ID: "toolbar_result"})
	// ANCHOR_END: demo
	return nil
}

func toolbarStickyDemo(p *tgframe.Params) error {
	// ANCHOR: sticky
	bar := tgcomp.Toolbar(p.Main, &tgcomp.ToolbarConf{
		ID:      "sticky_toolbar",
		Sticky:  true,
		Justify: tgcomp.ToolbarJustifyBetween,
	})

	tgcomp.Subtitle(bar, "Report")
	tgcomp.Button(bar, "Export", &tgcomp.ButtonConf{ID: "toolbar_export"})

	// Enough rows to scroll, so the row above stays put while they pass
	// under it.
	for i := range 40 {
		tgcomp.Text(p.Main, fmt.Sprintf("row-%d", i))
	}
	// ANCHOR_END: sticky
	return nil
}

func toolbarDialogDemo(p *tgframe.Params) error {
	// ANCHOR: dialog
	d := tgcomp.Dialog(p.Main, "Rows", &tgcomp.DialogConf{ID: "toolbar_dialog"})

	if tgcomp.Button(p.Main, "Open", &tgcomp.ButtonConf{ID: "toolbar_dialog_open"}) {
		d.Open()
	}

	d.With(func(c *tgframe.Container) {
		bar := tgcomp.Toolbar(c, &tgcomp.ToolbarConf{
			ID:     "dialog_toolbar",
			Sticky: true,
		})
		tgcomp.Button(bar, "Export", &tgcomp.ButtonConf{ID: "toolbar_dialog_export"})

		// A dialog scrolls its own body, so the row sticks to the top of
		// that -- below the dialog's header rather than under it.
		for i := range 30 {
			tgcomp.Text(c, fmt.Sprintf("dialog-row-%d", i))
		}
	})
	// ANCHOR_END: dialog
	return nil
}
