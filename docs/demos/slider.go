package demos

import (
	"fmt"

	"github.com/voilelab/toolgui/toolgui/tgcomp"
	"github.com/voilelab/toolgui/toolgui/tgcomp/tcinput"
	"github.com/voilelab/toolgui/toolgui/tgframe"
)

func sliderDemo(p *tgframe.Params) error {
	// ANCHOR: demo
	sliderValue := tgcomp.Slider(p.Main, "Slider",
		(&tcinput.SliderConf[int64]{}).SetMin(0).SetMax(100).SetStep(10).
			SetDefault(50))

	// A slider always sits somewhere, so there is always a value.
	tgcomp.Text(p.Main, fmt.Sprint("Value: ", sliderValue),
		&tgcomp.TextConf{ID: "slider_result"})
	// ANCHOR_END: demo
	return nil
}
