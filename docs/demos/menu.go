package demos

import (
	"github.com/voilelab/toolgui/toolgui/tgcomp"
	"github.com/voilelab/toolgui/toolgui/tgframe"
)

func menuDemo(p *tgframe.Params) error {
	// ANCHOR: demo
	items := []string{"Rename", "Duplicate", "Delete"}
	picked := tgcomp.Menu(p.Main, "Actions", items)

	action := "none"
	if picked != nil {
		action = items[*picked]
	}

	tgcomp.Text(p.Main, "Action: "+action,
		&tgcomp.TextConf{ID: "menu_result"})
	// ANCHOR_END: demo
	return nil
}
