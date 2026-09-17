package demos

import (
	"time"

	"github.com/voilelab/toolgui/toolgui/tgcomp"
	"github.com/voilelab/toolgui/toolgui/tgframe"
)

func toastDemo(p *tgframe.Params) error {
	// ANCHOR: demo
	if tgcomp.Button(p.Main, "Save") {
		tgcomp.Toast(p.Main, "Saved to disk",
			&tgcomp.ToastConf{Icon: "✅"})
		tgcomp.Toast(p.Main, "Two rows changed",
			&tgcomp.ToastConf{Icon: "📝", Duration: 10 * time.Second})
	}
	// ANCHOR_END: demo
	return nil
}
