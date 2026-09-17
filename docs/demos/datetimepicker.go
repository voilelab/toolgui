package demos

import (
	"github.com/voilelab/toolgui/toolgui/tgcomp"
	"github.com/voilelab/toolgui/toolgui/tgframe"
)

func dateTimePickerDemo(p *tgframe.Params) error {
	// ANCHOR: demo
	datetimeValue := tgcomp.DateTimePicker(p.Main, "DateTimePicker")
	val := ""
	if datetimeValue != nil {
		val = datetimeValue.Format("2006-01-02 15:04")
	}

	tgcomp.Text(p.Main, "Value: "+val,
		&tgcomp.TextConf{ID: "datetimepicker_result"})
	// ANCHOR_END: demo
	return nil
}
