package demos

import (
	"github.com/voilelab/toolgui/toolgui/tgcomp"
	"github.com/voilelab/toolgui/toolgui/tgframe"
)

func subtitleDemo(p *tgframe.Params) error {
	// ANCHOR: demo
	tgcomp.Subtitle(p.Main, "Subtitle")
	// ANCHOR_END: demo
	return nil
}
