package demos

import (
	"errors"

	"github.com/voilelab/toolgui/toolgui/tgcomp"
	"github.com/voilelab/toolgui/toolgui/tgframe"
)

func errorDemo(p *tgframe.Params) error {
	// ANCHOR: demo
	if tgcomp.Button(p.Main, "Show error") {
		return errors.New("new error")
	}
	// ANCHOR_END: demo
	return nil
}

func panicDemo(p *tgframe.Params) error {
	// ANCHOR: panic
	if tgcomp.Button(p.Main, "Show panic") {
		panic("show panic")
	}
	// ANCHOR_END: panic
	return nil
}
