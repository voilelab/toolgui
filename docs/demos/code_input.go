package demos

import (
	"github.com/voilelab/toolgui/toolgui/tgcomp"
	"github.com/voilelab/toolgui/toolgui/tgframe"
)

func codeInputDemo(p *tgframe.Params) error {
	// ANCHOR: demo
	code := tgcomp.CodeInput(p.Main, "Code",
		&tgcomp.CodeInputConf{
			Language: "python",
			Height:   6,
			Default:  "def hello(name):\n    return f\"Hello, {name}!\"\n",
		})
	tgcomp.Code(p.Main, code,
		&tgcomp.CodeConf{ID: "code_input_result", Language: "python"})
	// ANCHOR_END: demo
	return nil
}
