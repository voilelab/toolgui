package demos

import (
	"fmt"

	"github.com/voilelab/toolgui/toolgui/tgcomp"
	"github.com/voilelab/toolgui/toolgui/tgframe"
)

func progressBarDemo(p *tgframe.Params) error {
	// ANCHOR: demo
	pct := p.State.Default("misc_progress", 30)
	if tgcomp.Button(p.Main, "+10%") {
		*pct += 10
		if *pct > 100 {
			*pct = 0
		}
	}

	tgcomp.ProgressBar(p.Main, *pct,
		fmt.Sprintf("progress_bar: %d%%", *pct),
		&tgcomp.ProgressBarConf{ID: "misc_progress"})
	// ANCHOR_END: demo
	return nil
}
