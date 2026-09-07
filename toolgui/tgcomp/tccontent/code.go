package tccontent

import (
	"github.com/voilelab/toolgui/toolgui/tgframe"
)

var _ tgframe.Component = &codeComponent{}
var codeComponentName = "code_component"

type codeComponent struct {
	*tgframe.BaseComponent
	Code string `json:"code"`
	Lang string `json:"lang"`
}

func newCodeComponent(code string) *codeComponent {
	return &codeComponent{
		BaseComponent: &tgframe.BaseComponent{
			Name: codeComponentName,
		},
		Code: code,
	}
}

// CodeConf provide extra config for Code Component.
type CodeConf struct {
	tgframe.Base

	// Language is language of code block, leave empty to use `go`
	Language string
}

// Code create a code block with syntax highlight.
func Code(c *tgframe.Container, code string, conf ...*CodeConf) {
	cf := tgframe.OneConf("Code", conf)

	comp := newCodeComponent(code)
	tgframe.SetConfID(comp, cf)

	if cf.Language != "" {
		comp.Lang = cf.Language
	} else {
		comp.Lang = "go"
	}

	c.AddComponent(comp)
}
