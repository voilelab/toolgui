package tcinput

import (
	"github.com/voilelab/toolgui/toolgui/tgcomp/tcutil"
	"github.com/voilelab/toolgui/toolgui/tgframe"
)

var _ tgframe.Component = &textboxComponent{}
var textboxComponentName = "textbox_component"

type textboxComponent struct {
	*tgframe.BaseComponent
	Label       string       `json:"label"`
	MaxLength   int          `json:"max_length"`
	Placeholder string       `json:"placeholder"`
	Password    bool         `json:"password"`
	Disabled    bool         `json:"disabled"`
	Default     string       `json:"default"`
	Color       tcutil.Color `json:"color"`
}

func newTextboxComponent(label string) *textboxComponent {
	return &textboxComponent{
		BaseComponent: &tgframe.BaseComponent{
			Name: textboxComponentName,
			ID:   tcutil.NormalID(textboxComponentName, label),
		},
		Label: label,
	}
}

// TextboxConf is the configuration for the Textbox component
type TextboxConf struct {
	// Placeholder text to display in the textbox.
	Placeholder string

	// Maximum number of characters allowed in the textbox.
	// If 0, there is no character limit.
	MaxLength int

	// Indicates whether the textbox should mask input as asterisks.
	Password bool

	// Indicates whether the textbox should be disabled.
	Disabled bool

	// Default value of the textbox.
	Default string

	// Color defines the color of the textbox
	Color tcutil.Color

	// ID is the unique identifier for this textbox component
	ID string
}

// Textbox create a textbox and return its value.
func Textbox(s *tgframe.State, c *tgframe.Container, label string) string {
	return TextboxWithConf(s, c, label, nil)
}

// TextboxWithConf create a textbox and return its value.
func TextboxWithConf(s *tgframe.State, c *tgframe.Container, label string, conf *TextboxConf) string {
	if conf == nil {
		conf = &TextboxConf{}
	}

	comp := newTextboxComponent(label)
	comp.Placeholder = conf.Placeholder
	comp.MaxLength = conf.MaxLength
	comp.Password = conf.Password
	comp.Disabled = conf.Disabled
	comp.Color = conf.Color
	comp.Default = conf.Default
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
