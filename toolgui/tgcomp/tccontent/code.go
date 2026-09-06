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
	// Language is language of code block, leave empty to use `go`
	Language string

	// ID is id of the component
	ID string
}

// Code create a code block with syntax highlight.
func Code(c *tgframe.Container, code string) {
	CodeWithConf(c, code, nil)
}

// CodeWithConf create a code block with syntax highlight.
func CodeWithConf(c *tgframe.Container, code string, conf *CodeConf) {
	comp := newCodeComponent(code)
	if conf == nil {
		conf = &CodeConf{}
	}

	if conf.ID != "" {
		comp.SetID(conf.ID)
	}

	if conf.Language != "" {
		comp.Lang = conf.Language
	} else {
		comp.Lang = "go"
	}

	c.AddComponent(comp)
}
