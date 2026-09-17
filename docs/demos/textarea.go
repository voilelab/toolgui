package demos

import (
	"github.com/voilelab/toolgui/toolgui/tgcomp"
	"github.com/voilelab/toolgui/toolgui/tgcomp/tcutil"
	"github.com/voilelab/toolgui/toolgui/tgframe"
)

func textareaDemo(p *tgframe.Params) error {
	// ANCHOR: demo
	textareaValue := tgcomp.Textarea(p.Main, "Textarea",
		&tgcomp.TextareaConf{
			Height: 5,
			Color:  tcutil.ColorWarning,
		})
	tgcomp.Text(p.Main, "Value: "+textareaValue,
		&tgcomp.TextConf{ID: "textarea_result"})
	// ANCHOR_END: demo
	return nil
}
