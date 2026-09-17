package demos

import (
	"github.com/voilelab/toolgui/toolgui/tgcomp"
	"github.com/voilelab/toolgui/toolgui/tgcomp/tcutil"
	"github.com/voilelab/toolgui/toolgui/tgframe"
)

func textboxDemo(p *tgframe.Params) error {
	// ANCHOR: demo
	textboxValue := tgcomp.Textbox(p.Main, "Textbox", &tgcomp.TextboxConf{
		Placeholder: "input the value here",
		Color:       tcutil.ColorInfo,
	})
	tgcomp.Text(p.Main, "Value: "+textboxValue,
		&tgcomp.TextConf{ID: "textbox_result"})
	// ANCHOR_END: demo
	return nil
}
