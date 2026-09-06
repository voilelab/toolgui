package tccontent

import (
	"github.com/voilelab/toolgui/toolgui/tgframe"
)

var _ tgframe.Component = &titleComponent{}
var titleComponentName = "title_component"

type titleComponent struct {
	*tgframe.BaseComponent
	Text string `json:"text"`
}

func newTitleComponent(text string) *titleComponent {
	return &titleComponent{
		BaseComponent: &tgframe.BaseComponent{
			Name: titleComponentName,
		},
		Text: text,
	}
}

// Title show a title.
func Title(c *tgframe.Container, text string) {
	c.AddComponent(newTitleComponent(text))
}

// TitleWithID create a title component with a user specific id.
func TitleWithID(c *tgframe.Container, text string, id string) {
	comp := newTitleComponent(text)
	comp.SetID(id)
	c.AddComponent(comp)
}
