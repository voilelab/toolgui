package demos

import (
	"context"
	"fmt"
	"iter"
	"strings"
	"time"

	"github.com/voilelab/toolgui/toolgui/tgcomp"
	"github.com/voilelab/toolgui/toolgui/tgframe"
)

func writeStreamDemo(p *tgframe.Params) error {
	// ANCHOR: demo
	if tgcomp.Button(p.Main, "Stream a reply") {
		text, err := tgcomp.WriteStream(p.Main, fakeReply(p.Context))
		if err != nil {
			return err
		}

		tgcomp.Caption(p.Main, fmt.Sprintf("%d words", len(strings.Fields(text))))
	}
	// ANCHOR_END: demo
	return nil
}

// ANCHOR: reply
// fakeReply stands in for a model's reply: a word at a time.
func fakeReply(ctx context.Context) iter.Seq2[string, error] {
	const reply = "**WriteStream** writes the chunks into one Markdown " +
		"as they arrive, and returns the whole text once the stream ends."

	return func(yield func(string, error) bool) {
		for _, word := range strings.SplitAfter(reply, " ") {
			select {
			case <-ctx.Done():
				return
			case <-time.After(80 * time.Millisecond):
			}

			if !yield(word, nil) {
				return
			}
		}
	}
}

// ANCHOR_END: reply
