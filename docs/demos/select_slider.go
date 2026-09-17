package demos

import (
	"github.com/voilelab/toolgui/toolgui/tgcomp"
	"github.com/voilelab/toolgui/toolgui/tgframe"
)

func selectSliderDemo(p *tgframe.Params) error {
	// ANCHOR: demo
	sizes := []string{"S", "M", "L"}
	selIdx := tgcomp.SelectSlider(p.Main, "SelectSlider", sizes)

	tgcomp.Text(p.Main, "Value: "+sizes[selIdx],
		&tgcomp.TextConf{ID: "select_slider_result"})
	// ANCHOR_END: demo
	return nil
}
