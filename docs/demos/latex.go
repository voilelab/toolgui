package demos

import (
	"github.com/voilelab/toolgui/toolgui/tgcomp"
	"github.com/voilelab/toolgui/toolgui/tgframe"
)

func latexDemo(p *tgframe.Params) error {
	// ANCHOR: demo
	tgcomp.Latex(p.Main, "E = mc^2")
	// ANCHOR_END: demo
	return nil
}
