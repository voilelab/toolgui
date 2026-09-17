package demos

import (
	"github.com/voilelab/toolgui/toolgui/tgcomp"
	"github.com/voilelab/toolgui/toolgui/tgframe"
)

func titleDemo(p *tgframe.Params) error {
	// ANCHOR: demo
	tgcomp.Title(p.Main, "Title")
	// ANCHOR_END: demo
	return nil
}
