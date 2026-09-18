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
	numberValue, ok := tgcomp.Number(p.Main, "Number",
		(&tcinput.NumberConf[float64]{
			Placeholder: "input the value here",
			Color:       tcutil.ColorSuccess,
			Default:     10,
		}).SetMin(10).SetMax(20).SetStep(2))

	// Type 123 and the box goes red while the value here reads 20: out of
	// range, what arrives is pulled to the bound rather than left on the
	// last one that was inside it.
	tgcomp.Text(p.Main, fmt.Sprint("Value: ", numberValue),
		&tgcomp.TextConf{ID: "number_result"})

	// 20 is a fine number to show, but nobody typed it, so it is not a
	// number to save. ok is what tells the two apart.
	if tgcomp.Button(p.Main, "Save number",
		&tgcomp.ButtonConf{ID: "save_number"}) {
		if ok {
			tgcomp.Text(p.Main, fmt.Sprint("Saved: ", numberValue),
				&tgcomp.TextConf{ID: "number_saved"})
		} else {
			tgcomp.Text(p.Main, "Not saved: enter a value between 10 and 20",
				&tgcomp.TextConf{ID: "number_saved"})
		}
	}
	// ANCHOR_END: demo
	return nil
}
