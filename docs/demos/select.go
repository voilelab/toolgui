package demos

import (
	"fmt"

	"github.com/voilelab/toolgui/toolgui/tgcomp"
	"github.com/voilelab/toolgui/toolgui/tgframe"
)

func selectDemo(p *tgframe.Params) error {
	// ANCHOR: demo
	selIdx := tgcomp.Select(p.Main, "Select", []string{"Value1", "Value2"})

	selItem := ""
	if selIdx != nil {
		selItem = fmt.Sprintf("Value%d", (*selIdx)+1)
	}

	tgcomp.Text(p.Main, "Value: "+selItem,
		&tgcomp.TextConf{ID: "select_result"})
	// ANCHOR_END: demo
	return nil
}
