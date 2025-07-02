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
	// Height is the height of the textarea. default value is 3.
	Height int `json:"height"`

	// Default is the default value of the textarea.
	Default string `json:"default"`

	// Color defines the color of the textarea
	Color tcutil.Color

	// ID is the unique identifier for this textarea component
	ID string
}

// Textarea create a textarea and return its value.
func Textarea(s *tgframe.State, c *tgframe.Container, label string) string {
	return TextareaWithConf(s, c, label, nil)
}

// TextareaWithConf create a textarea and return its value.
func TextareaWithConf(s *tgframe.State, c *tgframe.Container, label string, conf *TextareaConf) string {
	if conf == nil {
		conf = &TextareaConf{}
	}

	comp := newTextareaComponent(label)
	comp.Height = conf.Height
	if comp.Height == 0 {
		comp.Height = 3
	}

	comp.Default = conf.Default
	comp.Color = conf.Color

	if conf.ID != "" {
		comp.SetID(conf.ID)
	}

	c.AddComponent(comp)
	val := s.GetString(comp.ID)
	if val == nil {
		return comp.Default
	}

	return *val
}
