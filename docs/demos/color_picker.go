package demos

import (
	"github.com/voilelab/toolgui/toolgui/tgcomp"
	"github.com/voilelab/toolgui/toolgui/tgframe"
)

func colorPickerDemo(p *tgframe.Params) error {
	// ANCHOR: demo
	color := tgcomp.ColorPicker(p.Main, "ColorPicker",
		&tgcomp.ColorPickerConf{Default: "#ff3860"})

	tgcomp.Text(p.Main, "Value: "+color,
		&tgcomp.TextConf{ID: "color_picker_result"})
	// ANCHOR_END: demo
	return nil
}
