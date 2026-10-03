package demos

import (
	"github.com/voilelab/toolgui/toolgui/tgcomp"
	"github.com/voilelab/toolgui/toolgui/tgframe"
)

func chatMessageDemo(p *tgframe.Params) error {
	// ANCHOR: demo
	tgcomp.Markdown(tgcomp.ChatMessage(p.Main, "user"), "What can ChatMessage hold?")

	reply := tgcomp.ChatMessage(p.Main, "assistant")
	tgcomp.Markdown(reply, "Any component, for example a **metric**:")
	tgcomp.Metric(reply, "Tokens", "1,024")

	tgcomp.Text(tgcomp.ChatMessage(p.Main, "tool",
		&tgcomp.ChatMessageConf{Avatar: "🔧"}), "search(\"toolgui\") -> 3 results")
	// ANCHOR_END: demo
	return nil
}

func chatMessageStreamDemo(p *tgframe.Params) error {
	// ANCHOR: stream
	if tgcomp.Button(p.Main, "Ask") {
		tgcomp.Text(tgcomp.ChatMessage(p.Main, "user"), "Tell me about WriteStream.")

		reply := tgcomp.ChatMessage(p.Main, "assistant")
		if _, err := tgcomp.WriteStream(reply, fakeReply(p.Context)); err != nil {
			return err
		}
	}
	// ANCHOR_END: stream
	return nil
}
