package demos

import (
	"github.com/voilelab/toolgui/toolgui/tgcomp"
	"github.com/voilelab/toolgui/toolgui/tgframe"
)

// A shortcode expands wherever text is decoration, and stays literal wherever
// it is the thing being shown, which is why the book walks the two halves of
// this example through separately.
func emojiDemo(p *tgframe.Params) error {
	// ANCHOR: text
	tgcomp.Text(p.Main, "Shipped it :tada:")
	// ANCHOR_END: text

	// ANCHOR: markdown
	tgcomp.Markdown(p.Main, "A `:tada:` in code stays as written.")
	// ANCHOR_END: markdown
	return nil
}
