package demos

import (
	"github.com/voilelab/toolgui/toolgui/tgcomp"
	"github.com/voilelab/toolgui/toolgui/tgframe"
)

func tabDemo(p *tgframe.Params) error {
	// ANCHOR: demo
	tab1, tab2 := tgcomp.Tab2(p.Main, "tab1", "tab2")
	tgcomp.Text(tab1, "A tab!")
	tgcomp.Text(tab2, "B tab!")
	// ANCHOR_END: demo
	return nil
}
