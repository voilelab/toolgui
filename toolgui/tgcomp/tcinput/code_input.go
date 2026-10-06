package tcinput

import (
	"github.com/voilelab/toolgui/toolgui/tgcomp/tcutil"
	"github.com/voilelab/toolgui/toolgui/tgframe"
)

var _ tgframe.Component = &codeInputComponent{}
var codeInputComponentName = "code_input_component"

type codeInputComponent struct {
	*tgframe.BaseComponent
	Label    string `json:"label"`
	Lang     string `json:"lang"`
	Height   int    `json:"height"`
	Default  string `json:"default"`
	ResetKey string `json:"reset_key"`
}

func newCodeInputComponent(label string) *codeInputComponent {
	return &codeInputComponent{
		BaseComponent: &tgframe.BaseComponent{
			Name: codeInputComponentName,
			ID:   tcutil.NormalID(codeInputComponentName, label),
		},
		Label: label,
	}
}

// CodeInputConf is the configuration for a code input.
type CodeInputConf struct {
	tgframe.Base

	// Language is the language to highlight, leave empty to use `go`.
	// Supported: go, python, javascript, typescript, jsx, tsx, json, sql,
	// html, css, markdown, yaml, shell. "text" or any other value
	// leaves the code unhighlighted.
	Language string

	// Height is the number of lines shown. default value is 10.
	Height int

	// Default is the default value of the code input.
	Default string

	// ResetKey drops the app user's input and restores Default whenever it
	// changes, e.g. a hash of the file the code was filled from.
	ResetKey string
}

// CodeInput create a code editor with syntax highlight and return its value.
// The value is sent when the editor loses focus or on Ctrl/Cmd+Enter.
func CodeInput(c *tgframe.Container, label string, conf ...*CodeInputConf) string {
	cf := tgframe.OneConf("CodeInput", conf)

	comp := newCodeInputComponent(label)
	comp.Lang = cf.Language
	if comp.Lang == "" {
		comp.Lang = "go"
	}
	comp.Height = cf.Height
	if comp.Height == 0 {
		comp.Height = 10
	}

	comp.Default = cf.Default
	comp.ResetKey = cf.ResetKey
	tgframe.SetConfIDIn(c, comp, cf)

	c.AddComponent(comp)
	resetOnKeyChange(c.State, comp.ID, comp.ResetKey)
	val, ok := c.State.Get[string](comp.ID)
	if !ok {
		return comp.Default
	}

	return val
}
