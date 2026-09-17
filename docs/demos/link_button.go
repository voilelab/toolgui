package demos

import (
	"github.com/voilelab/toolgui/toolgui/tgcomp"
	"github.com/voilelab/toolgui/toolgui/tgframe"
)

func linkButtonDemo(p *tgframe.Params) error {
	// ANCHOR: demo
	tgcomp.LinkButton(p.Main, "Link Button",
		"https://www.example.com/")
	// ANCHOR_END: demo
	return nil
}
