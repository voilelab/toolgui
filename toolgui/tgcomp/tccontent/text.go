package tccontent

import (
	"github.com/voilelab/toolgui/toolgui/tgframe"
)

var _ tgframe.Component = &textComponent{}
var textComponentName = "text_component"

type textComponent struct {
	*tgframe.BaseComponent
	Text string `json:"text"`
}

func newTextComponent(text string) *textComponent {
	return &textComponent{
		BaseComponent: &tgframe.BaseComponent{
			Name: textComponentName,
		},
		Text: text,
	}
}

// TextConf is the configuration for the Text component.
type TextConf struct {
	// ID is the unique identifier for this text component.
	ID string
}

// Text show a text. conf may be nil.
func Text(c *tgframe.Container, text string, conf *TextConf) {
	comp := newTextComponent(text)
	if conf != nil && conf.ID != "" {
		comp.SetID(conf.ID)
	}

	c.AddComponent(comp)
}
