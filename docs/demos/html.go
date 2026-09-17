package demos

import (
	"github.com/voilelab/toolgui/toolgui/tgcomp"
	"github.com/voilelab/toolgui/toolgui/tgframe"
)

func htmlDemo(p *tgframe.Params) error {
	// ANCHOR: demo
	tgcomp.HTML(p.Main,
		"<b>Hello world gen by html component</b>")
	// ANCHOR_END: demo
	return nil
}
