package demos

import (
	"github.com/voilelab/toolgui/toolgui/tgcomp"
	"github.com/voilelab/toolgui/toolgui/tgframe"
)

func popoverDemo(p *tgframe.Params) error {
	// ANCHOR: demo
	pop := tgcomp.Popover(p.Main, "Advanced options")
	tgcomp.Checkbox(pop, "Show hidden columns")
	if tgcomp.Button(pop, "Reset options") {
		tgcomp.Text(p.Main, "Options reset.",
			&tgcomp.TextConf{ID: "popover_reset"})
	}
	// ANCHOR_END: demo
	return nil
}
