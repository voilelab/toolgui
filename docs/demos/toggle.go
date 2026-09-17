package demos

import (
	"fmt"

	"github.com/voilelab/toolgui/toolgui/tgcomp"
	"github.com/voilelab/toolgui/toolgui/tgframe"
)

func toggleDemo(p *tgframe.Params) error {
	// ANCHOR: demo
	toggleValue := tgcomp.Toggle(p.Main, "Toggle")
	tgcomp.Text(p.Main, fmt.Sprint("Value: ", toggleValue),
		&tgcomp.TextConf{ID: "toggle_result"})

	onValue := tgcomp.Toggle(p.Main, "Toggle default on",
		&tgcomp.ToggleConf{Default: true})
	tgcomp.Text(p.Main, fmt.Sprint("Default: ", onValue),
		&tgcomp.TextConf{ID: "toggle_default_result"})
	// ANCHOR_END: demo
	return nil
}
