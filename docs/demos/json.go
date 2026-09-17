package demos

import (
	"github.com/voilelab/toolgui/toolgui/tgcomp"
	"github.com/voilelab/toolgui/toolgui/tgframe"
)

func jsonDemo(p *tgframe.Params) error {
	// ANCHOR: demo
	type DemoJSONHeader struct {
		Type int
	}

	type DemoJSON struct {
		Header   DemoJSONHeader
		IntValue int
		URL      string
		IsOk     bool
	}

	tgcomp.JSON(p.Main, &DemoJSON{})
	// ANCHOR_END: demo
	return nil
}
