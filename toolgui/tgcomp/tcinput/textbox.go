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
	tgframe.Base

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
}

// Textbox create a textbox and return its value.
func Textbox(c *tgframe.Container, label string, conf ...*TextboxConf) string {
	cf := tgframe.OneConf(conf)

	comp := newTextboxComponent(label)
	comp.Placeholder = cf.Placeholder
	comp.MaxLength = cf.MaxLength
	comp.Password = cf.Password
	comp.Disabled = cf.Disabled
	comp.Color = cf.Color
	comp.Default = cf.Default
	tgframe.SetConfID(comp, cf)

	c.AddComponent(comp)
	val := c.State.GetString(comp.ID)
	if val == nil {
		return comp.Default
	}

	return *val
}
