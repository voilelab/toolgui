package demos

import (
	"fmt"

	"github.com/voilelab/toolgui/toolgui/tgcomp"
	"github.com/voilelab/toolgui/toolgui/tgframe"
)

func checkboxDemo(p *tgframe.Params) error {
	// ANCHOR: demo
	checkboxValue := tgcomp.Checkbox(p.Main, "Checkbox")
	tgcomp.Text(p.Main, fmt.Sprint("Value: ", checkboxValue),
		&tgcomp.TextConf{ID: "checkbox_result"})

	onValue := tgcomp.Checkbox(p.Main, "Checkbox default on",
		&tgcomp.CheckboxConf{Default: true})
	tgcomp.Text(p.Main, fmt.Sprint("Default: ", onValue),
		&tgcomp.TextConf{ID: "checkbox_default_result"})
	// ANCHOR_END: demo
	return nil
}
