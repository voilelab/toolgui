package demos

import (
	"time"

	"github.com/voilelab/toolgui/toolgui/tgcomp"
	"github.com/voilelab/toolgui/toolgui/tgframe"
)

func statusDemo(p *tgframe.Params) error {
	// ANCHOR: demo
	if tgcomp.Button(p.Main, "Import three files") {
		status := tgcomp.Status(p.Main, "Importing…",
			&tgcomp.StatusConf{Expanded: true})

		for _, name := range []string{"one.csv", "two.csv", "three.csv"} {
			status.Write(name)
			time.Sleep(time.Second)
		}

		status.Complete("Imported 3 files")
	}
	// ANCHOR_END: demo
	return nil
}
