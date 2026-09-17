package demos

import (
	"time"

	"github.com/voilelab/toolgui/toolgui/tgcomp"
	"github.com/voilelab/toolgui/toolgui/tgframe"
)

func emptyDemo(p *tgframe.Params) error {
	// ANCHOR: demo
	slot := tgcomp.Empty(p.Main, &tgcomp.EmptyConf{ID: "query_result"})
	slot.With(func(c *tgframe.Container) {
		tgcomp.Text(c, "No query yet.")
	})

	if tgcomp.Button(p.Main, "Run a slow query") {
		slot.With(func(c *tgframe.Container) {
			tgcomp.Text(c, "Querying…")
		})

		time.Sleep(3 * time.Second)

		slot.With(func(c *tgframe.Container) {
			tgcomp.Table(c,
				[]string{"table", "rows"},
				[][]string{{"users", "1289"}, {"orders", "4021"}})
		})
	}
	// ANCHOR_END: demo
	return nil
}
