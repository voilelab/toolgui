package demos

import (
	"github.com/voilelab/toolgui/toolgui/tgcomp"
	"github.com/voilelab/toolgui/toolgui/tgcomp/tcutil"
	"github.com/voilelab/toolgui/toolgui/tgframe"
)

func badgeDemo(p *tgframe.Params) error {
	// ANCHOR: demo
	tgcomp.Badge(p.Main, "Badge")
	tgcomp.Badge(p.Main, "Shipped",
		&tgcomp.BadgeConf{Color: tcutil.ColorSuccess})
	// ANCHOR_END: demo
	return nil
}
