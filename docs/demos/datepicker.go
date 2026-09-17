package demos

import (
	"github.com/voilelab/toolgui/toolgui/tgcomp"
	"github.com/voilelab/toolgui/toolgui/tgframe"
)

func datePickerDemo(p *tgframe.Params) error {
	// ANCHOR: demo
	dateValue := tgcomp.DatePicker(p.Main, "DatePicker")
	val := ""
	if dateValue != nil {
		val = dateValue.Format("2006-01-02")
	}

	tgcomp.Text(p.Main, "Value: "+val,
		&tgcomp.TextConf{ID: "datepicker_result"})
	// ANCHOR_END: demo
	return nil
}
