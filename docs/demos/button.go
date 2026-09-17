package demos

import (
	"fmt"

	"github.com/voilelab/toolgui/toolgui/tgcomp"
	"github.com/voilelab/toolgui/toolgui/tgframe"
)

func buttonDemo(p *tgframe.Params) error {
	// ANCHOR: demo
	btnClicked := tgcomp.Button(p.Main, "button")
	tgcomp.Text(p.Main, fmt.Sprint("Value: ", btnClicked),
		&tgcomp.TextConf{ID: "button_result"})
	// ANCHOR_END: demo
	return nil
}
