package demos

import (
	"github.com/voilelab/toolgui/toolgui/tgcomp"
	"github.com/voilelab/toolgui/toolgui/tgframe"
)

func tableDemo(p *tgframe.Params) error {
	// ANCHOR: demo
	tgcomp.Table(p.Main, []string{"a", "b"},
		[][]string{{"1", "2"}, {"3", "4"}})
	// ANCHOR_END: demo
	return nil
}
