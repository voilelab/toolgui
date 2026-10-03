package demos

import (
	"context"
	"iter"
	"strings"
	"time"

	"github.com/voilelab/toolgui/toolgui/tgcomp"
	"github.com/voilelab/toolgui/toolgui/tgframe"
)

// ANCHOR: turn
type chatTurn struct {
	Role string
	Text string
}

// ANCHOR_END: turn

func chatInputDemo(p *tgframe.Params) error {
	// ANCHOR: demo
	history := p.State.Default("chat_history", []chatTurn{})
	msg, sent := tgcomp.ChatInput(p.Main, "Say something")

	for _, turn := range *history {
		tgcomp.Markdown(tgcomp.ChatMessage(p.Main, turn.Role), turn.Text)
	}

	if sent {
		tgcomp.Markdown(tgcomp.ChatMessage(p.Main, "user"), msg)

		reply, err := tgcomp.WriteStream(
			tgcomp.ChatMessage(p.Main, "assistant"), echo(p.Context, msg))
		if err != nil {
			return err
		}

		*history = append(*history,
			chatTurn{Role: "user", Text: msg},
			chatTurn{Role: "assistant", Text: reply})
	}
	// ANCHOR_END: demo
	return nil
}

// ANCHOR: echo
// echo stands in for a model: it answers with the message, a word at a time.
func echo(ctx context.Context, msg string) iter.Seq2[string, error] {
	return func(yield func(string, error) bool) {
		for _, word := range strings.SplitAfter("You said: "+msg, " ") {
			select {
			case <-ctx.Done():
				return
			case <-time.After(60 * time.Millisecond):
			}

			if !yield(word, nil) {
				return
			}
		}
	}
}

// ANCHOR_END: echo
