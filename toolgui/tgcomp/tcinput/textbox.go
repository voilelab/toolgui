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
	ResetKey    string       `json:"reset_key"`
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

	// ResetKey drops the app user's input and restores Default whenever it
	// changes, e.g. a hash of the file the text was filled from.
	ResetKey string

	// Color defines the color of the textbox
	Color tcutil.Color
}

// Textbox create a textbox and return its value.
func Textbox(c *tgframe.Container, label string, conf ...*TextboxConf) string {
	cf := tgframe.OneConf("Textbox", conf)

	comp := newTextboxComponent(label)
	comp.Placeholder = cf.Placeholder
	comp.MaxLength = cf.MaxLength
	comp.Password = cf.Password
	comp.Disabled = cf.Disabled
	comp.Color = cf.Color
	comp.Default = cf.Default
	comp.ResetKey = cf.ResetKey
	tgframe.SetConfID(comp, cf)

	c.AddComponent(comp)
	resetOnKeyChange(c.State, comp.ID, comp.ResetKey)
	val, ok := c.State.Get[string](comp.ID)
	if !ok {
		return comp.Default
	}

	return val
}

// resetOnKeyChange deletes the value under id when resetKey differs from the
// one seen last run. The first sighting only records it, so a value set
// before the first draw survives.
func resetOnKeyChange(s *tgframe.State, id, resetKey string) {
	keyID := id + "#reset_key"
	last, ok := s.Get[string](keyID)
	if ok && last == resetKey {
		return
	}

	if ok {
		s.Delete(id)
	}
	s.Set(keyID, resetKey)
}
