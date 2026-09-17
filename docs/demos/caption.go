package demos

import (
	"github.com/voilelab/toolgui/toolgui/tgcomp"
	"github.com/voilelab/toolgui/toolgui/tgframe"
)

func captionDemo(p *tgframe.Params) error {
	// ANCHOR: demo
	tgcomp.Caption(p.Main, "Caption")
	// ANCHOR_END: demo
	return nil
}
