package demos

import (
	"net/url"
	"strconv"

	"github.com/voilelab/toolgui/toolgui/tgcomp"
	"github.com/voilelab/toolgui/toolgui/tgframe"
)

func pageLinkDemo(p *tgframe.Params) error {
	// ANCHOR: demo
	// The query is untrusted: anything that is not a number is 0.
	n, err := strconv.Atoi(p.Query.Get("n"))
	if err != nil || n < 0 {
		n = 0
	}

	tgcomp.Text(p.Main, "n = "+strconv.Itoa(n))
	tgcomp.PageLink(p.Main, "Next", "page_link", url.Values{
		"n": {strconv.Itoa(n + 1)},
	})
	// ANCHOR_END: demo
	return nil
}
