package demos

import (
	"github.com/voilelab/toolgui/toolgui/tgcomp"
	"github.com/voilelab/toolgui/toolgui/tgcomp/tcutil"
	"github.com/voilelab/toolgui/toolgui/tgframe"
)

func messageDemo(p *tgframe.Params) error {
	// ANCHOR: demo
	tgcomp.Message(p.Main, "body of msg")

	tgcomp.Message(p.Main, "body of msg2",
		&tgcomp.MessageConf{
			Title: "danger!",
			Color: tcutil.ColorDanger,
		})
	// ANCHOR_END: demo
	return nil
}
