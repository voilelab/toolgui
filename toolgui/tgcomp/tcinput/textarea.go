package tcinput

import (
	"github.com/voilelab/toolgui/toolgui/tgcomp/tcutil"
	"github.com/voilelab/toolgui/toolgui/tgframe"
)

var _ tgframe.Component = &textareaComponent{}
var textareaComponentName = "textarea_component"

type textareaComponent struct {
	*tgframe.BaseComponent
	Label   string       `json:"label"`
	Height  int          `json:"height"`
	Default string       `json:"default"`
	Color   tcutil.Color `json:"color"`
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
	comp.Color = cf.Color
	tgframe.SetConfID(comp, cf)

	c.AddComponent(comp)
	val := c.State.GetString(comp.ID)
	if val == nil {
		return comp.Default
	}

	return *val
}
