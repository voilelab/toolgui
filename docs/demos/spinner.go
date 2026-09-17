package demos

import (
	"time"

	"github.com/voilelab/toolgui/toolgui/tgcomp"
	"github.com/voilelab/toolgui/toolgui/tgframe"
)

func spinnerDemo(p *tgframe.Params) error {
	// ANCHOR: demo
	if tgcomp.Button(p.Main, "Spin for three seconds") {
		stop := tgcomp.Spinner(p.Main, "Working…")
		time.Sleep(3 * time.Second)
		stop()

		tgcomp.Text(p.Main, "Done!")
	}
	// ANCHOR_END: demo
	return nil
}
