package demos

import (
	"github.com/voilelab/toolgui/toolgui/tgcomp"
	"github.com/voilelab/toolgui/toolgui/tgframe"
)

func textDemo(p *tgframe.Params) error {
	// ANCHOR: demo
	tgcomp.Text(p.Main, "Text")
	// ANCHOR_END: demo
	return nil
}

// Components are placed by position, so writing the same thing twice shows
// it twice.
func textDuplicateDemo(p *tgframe.Params) error {
	// ANCHOR: duplicate
	tgcomp.Text(p.Main, "written twice")
	tgcomp.Text(p.Main, "written twice")
	// ANCHOR_END: duplicate
	return nil
}
