package demos

import (
	"github.com/voilelab/toolgui/toolgui/tgcomp"
	"github.com/voilelab/toolgui/toolgui/tgframe"
)

func expandDemo(p *tgframe.Params) error {
	// ANCHOR: demo
	expand := tgcomp.Expand(p.Main, "Expand", true)
	tgcomp.Text(expand, "A expand!")
	// ANCHOR_END: demo
	return nil
}
