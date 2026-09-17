package demos

import (
	"fmt"

	"github.com/voilelab/toolgui/toolgui/tgcomp"
	"github.com/voilelab/toolgui/toolgui/tgcomp/tcinput"
	"github.com/voilelab/toolgui/toolgui/tgcomp/tcutil"
	"github.com/voilelab/toolgui/toolgui/tgframe"
)

func numberDemo(p *tgframe.Params) error {
	// ANCHOR: demo
	numberValue := tgcomp.Number(p.Main, "Number", (&tcinput.NumberConf[float64]{
		Placeholder: "input the value here",
		Color:       tcutil.ColorSuccess,
		Default:     10,
	}).SetMin(10).SetMax(20).SetStep(2))

	// Type 123 and the box goes red while the value here reads 20: out of
	// range, what arrives is pulled to the bound rather than left on the
	// last one that was inside it.
	tgcomp.Text(p.Main, fmt.Sprint("Value: ", numberValue),
		&tgcomp.TextConf{ID: "number_result"})

	// So a button pressed while the box is red cannot act on a number the
	// app user has already replaced.
	if tgcomp.Button(p.Main, "Save number",
		&tgcomp.ButtonConf{ID: "save_number"}) {
		tgcomp.Text(p.Main, fmt.Sprint("Saved: ", numberValue),
			&tgcomp.TextConf{ID: "number_saved"})
	}
	// ANCHOR_END: demo
	return nil
}
