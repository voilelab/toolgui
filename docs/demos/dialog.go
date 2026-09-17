package demos

import (
	"fmt"

	"github.com/voilelab/toolgui/toolgui/tgcomp"
	"github.com/voilelab/toolgui/toolgui/tgframe"
)

func dialogDemo(p *tgframe.Params) error {
	// ANCHOR: demo
	// Written before the dialog that opens it, on purpose: which of two
	// open dialogs is drawn on top follows the order they opened, not the
	// order the page writes them.
	why := tgcomp.Dialog(p.Main, "What deleting does")

	d := tgcomp.Dialog(p.Main, "Delete confirm")

	// The trigger is handled before With, which is what draws the body:
	// opening after it would open the dialog on an empty run.
	if tgcomp.Button(p.Main, "Delete") {
		d.Open()
	}

	d.With(func(c *tgframe.Container) {
		// Only reached while the dialog is open, so the count behind the
		// question is not looked up on every rerun.
		tgcomp.Text(c, fmt.Sprintf("Delete the %d selected rows?", 3))

		yes, no := tgcomp.EqColumn2(c, &tgcomp.ColumnConf{ID: "delete_confirm"})
		if tgcomp.Button(yes, "Yes, delete") {
			d.Close()
		}
		if tgcomp.Button(no, "Keep them") {
			d.Close()
		}

		if tgcomp.Button(c, "What does this do?") {
			why.Open()
		}

		// A popover inside a dialog: ESC reaches whichever was opened
		// last, so this closes before the dialog around it does.
		pop := tgcomp.Popover(c, "Delete options")
		tgcomp.Checkbox(pop, "Also delete the log")
	})

	why.With(func(c *tgframe.Container) {
		tgcomp.Text(c, "The rows are removed for good.")
		if tgcomp.Button(c, "Got it") {
			why.Close()
		}
	})
	// ANCHOR_END: demo
	return nil
}

func dialogSidebarDemo(p *tgframe.Params) error {
	// ANCHOR: sidebar
	// A dialog is a portal wherever it is written, so one declared in the
	// sidebar covers the whole window rather than the side column.
	sideDialog := tgcomp.Dialog(p.Sidebar, "From the sidebar",
		(&tgcomp.DialogConf{Width: tgcomp.DialogWidthMedium}).
			SetDismissible(false))
	if tgcomp.Button(p.Sidebar, "Open the sidebar dialog") {
		sideDialog.Open()
	}
	sideDialog.With(func(c *tgframe.Container) {
		tgcomp.Text(c, "Declared in the sidebar, shown over the page.")
		if tgcomp.Button(c, "Close the sidebar dialog") {
			sideDialog.Close()
		}
	})
	// ANCHOR_END: sidebar
	return nil
}
