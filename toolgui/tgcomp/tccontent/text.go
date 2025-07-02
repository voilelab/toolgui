package tccontent

import (
	"github.com/voilelab/toolgui/toolgui/tgcomp/tcutil"
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
			ID:   tcutil.NormalID(textComponentName, text),
		},
		Text: text,
	}
}

// Text show a text.
func Text(c *tgframe.Container, text string) {
	comp := newTextComponent(text)
	c.AddComponent(comp)
}

// TextWithID create a text component with a user specific id.
func TextWithID(c *tgframe.Container, text string, id string) {
	comp := newTextComponent(text)
	comp.SetID(id)
	c.AddComponent(comp)
}
