package demos

import (
	_ "embed"

	"github.com/voilelab/toolgui/toolgui/tgcomp"
	"github.com/voilelab/toolgui/toolgui/tgframe"
)

// code is this file, which is what Echo reads the lambda's lines out of.
// Every other demo prints its source through the anchors instead; this one is
// about the component that does it for itself.
//
//go:embed echo.go
var code string

func echoDemo(p *tgframe.Params) error {
	// ANCHOR: demo
	tgcomp.Echo(p.Main, code, func() {
		tgcomp.Text(p.Main, "hello echo")
	})
	// ANCHOR_END: demo
	return nil
}
