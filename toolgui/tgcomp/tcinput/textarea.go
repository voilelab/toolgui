package tcinput

import (
	"github.com/voilelab/toolgui/toolgui/tgcomp/tcutil"
	"github.com/voilelab/toolgui/toolgui/tgframe"
)

var _ tgframe.Component = &textareaComponent{}
var textareaComponentName = "textarea_component"

type textareaComponent struct {
	*tgframe.BaseComponent
	Label    string       `json:"label"`
	Height   int          `json:"height"`
	Default  string       `json:"default"`
	ResetKey string       `json:"reset_key"`
	Color    tcutil.Color `json:"color"`
}

func newTextareaComponent(label string) *textareaComponent {
	return &textareaComponent{
		BaseComponent: &tgframe.BaseComponent{
			Name: textareaComponentName,
			ID:   tcutil.NormalID(textareaComponentName, label),
		},
		Label: label,
	}
}

// TextareaConf is the configuration for a textarea.
type TextareaConf struct {
	tgframe.Base

	// Height is the height of the textarea. default value is 3.
	Height int

	// Default is the default value of the textarea.
	Default string

	// ResetKey drops the app user's input and restores Default whenever it
	// changes, e.g. a hash of the file the text was filled from.
	ResetKey string

	// Color defines the color of the textarea
	Color tcutil.Color
}

// Textarea create a textarea and return its value.
func Textarea(c *tgframe.Container, label string, conf ...*TextareaConf) string {
	cf := tgframe.OneConf("Textarea", conf)

	comp := newTextareaComponent(label)
	comp.Height = cf.Height
	if comp.Height == 0 {
		comp.Height = 3
	}

	comp.Default = cf.Default
	comp.ResetKey = cf.ResetKey
	comp.Color = cf.Color
	tgframe.SetConfID(comp, cf)

	c.AddComponent(comp)
	resetOnKeyChange(c.State, comp.ID, comp.ResetKey)
	val, ok := c.State.Get[string](comp.ID)
	if !ok {
		return comp.Default
	}

	return val
}
