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
	tgframe.Base
}

// Text show a text.
func Text(c *tgframe.Container, text string, conf ...*TextConf) {
	cf := tgframe.OneConf(conf)

	comp := newTextComponent(text)
	tgframe.SetConfID(comp, cf)
	c.AddComponent(comp)
}
