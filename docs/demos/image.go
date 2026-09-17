package demos

import (
	"github.com/voilelab/toolgui/toolgui/tgcomp"
	"github.com/voilelab/toolgui/toolgui/tgframe"
)

func imageDemo(p *tgframe.Params) error {
	// ANCHOR: demo
	tgcomp.Image(p.Main, "https://http.cat/100",
		&tgcomp.ImageConf{
			Width: "200px",
		})
	// ANCHOR_END: demo
	return nil
}
