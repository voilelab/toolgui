package demos

import (
	"github.com/voilelab/toolgui/toolgui/tgcomp"
	"github.com/voilelab/toolgui/toolgui/tgframe"
)

func boxDemo(p *tgframe.Params) error {
	// ANCHOR: demo
	box := tgcomp.Box(p.Main, &tgcomp.BoxConf{ID: "box"})
	tgcomp.Text(box, "A box!")
	// ANCHOR_END: demo
	return nil
}
