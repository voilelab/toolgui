package demos

import (
	"github.com/voilelab/toolgui/toolgui/tgcomp"
	"github.com/voilelab/toolgui/toolgui/tgframe"
)

func timePickerDemo(p *tgframe.Params) error {
	// ANCHOR: demo
	timeValue := tgcomp.TimePicker(p.Main, "TimePicker")
	val := ""
	if timeValue != nil {
		val = timeValue.Format("15:04")
	}

	tgcomp.Text(p.Main, "Value: "+val,
		&tgcomp.TextConf{ID: "timepicker_result"})
	// ANCHOR_END: demo
	return nil
}
