package demos

import (
	"github.com/voilelab/toolgui/toolgui/tgcomp"
	"github.com/voilelab/toolgui/toolgui/tgframe"
)

func dividerDemo(p *tgframe.Params) error {
	// ANCHOR: demo
	tgcomp.Divider(p.Main)
	// ANCHOR_END: demo
	return nil
}
