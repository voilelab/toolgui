package demos

import (
	"fmt"

	"github.com/voilelab/toolgui/toolgui/tgcomp"
	"github.com/voilelab/toolgui/toolgui/tgframe"
)

func radioDemo(p *tgframe.Params) error {
	// ANCHOR: demo
	selIdx := tgcomp.Radio(p.Main,
		"Radio", []string{"Value3", "Value4"})

	selItem := ""
	if selIdx != nil {
		selItem = fmt.Sprintf("Value%d", (*selIdx)+3)
	}

	tgcomp.Text(p.Main, "Value: "+selItem,
		&tgcomp.TextConf{ID: "radio_result"})
	// ANCHOR_END: demo
	return nil
}
