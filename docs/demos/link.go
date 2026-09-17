package demos

import (
	"github.com/voilelab/toolgui/toolgui/tgcomp"
	"github.com/voilelab/toolgui/toolgui/tgframe"
)

func linkDemo(p *tgframe.Params) error {
	// ANCHOR: demo
	tgcomp.Link(p.Main, "Link", "https://www.example.com/")
	// ANCHOR_END: demo
	return nil
}
