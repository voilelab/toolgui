package demos

import (
	"strings"

	"github.com/voilelab/toolgui/toolgui/tgcomp"
	"github.com/voilelab/toolgui/toolgui/tgframe"
)

func multiSelectDemo(p *tgframe.Params) error {
	// ANCHOR: demo
	items := []string{"Alpha", "Beta", "Gamma"}
	selIdxes := tgcomp.MultiSelect(p.Main, "MultiSelect", items,
		&tgcomp.MultiSelectConf{
			Placeholder:   "pick up to two",
			MaxSelections: 2,
		})

	selItems := []string{}
	for _, idx := range selIdxes {
		selItems = append(selItems, items[idx])
	}

	tgcomp.Text(p.Main, "Values: "+strings.Join(selItems, ", "),
		&tgcomp.TextConf{ID: "multiselect_result"})
	// ANCHOR_END: demo
	return nil
}
