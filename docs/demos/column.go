package demos

import (
	"fmt"

	"github.com/voilelab/toolgui/toolgui/tgcomp"
	"github.com/voilelab/toolgui/toolgui/tgframe"
)

func columnDemo(p *tgframe.Params) error {
	// ANCHOR: demo
	cols := tgcomp.Column(p.Main, 3, &tgcomp.ColumnConf{ID: "cols"})
	for i, col := range cols {
		tgcomp.Text(col, fmt.Sprintf("col-%d", i))
	}
	// ANCHOR_END: demo
	return nil
}
