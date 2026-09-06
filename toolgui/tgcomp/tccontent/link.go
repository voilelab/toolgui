package tccontent

import (
	"github.com/voilelab/toolgui/toolgui/tgframe"
)

var _ tgframe.Component = &linkComponent{}
var linkComponentName = "link_component"

type linkComponent struct {
	*tgframe.BaseComponent
	Text string `json:"text"`
	URL  string `json:"url"`
}

func newLinkComponent(text, url string) *linkComponent {
	return &linkComponent{
		BaseComponent: &tgframe.BaseComponent{
			Name: linkComponentName,
		},
		Text: text,
		URL:  url,
	}
}

// Link create a link component.
func Link(c *tgframe.Container, text, url string) {
	c.AddComponent(newLinkComponent(text, url))
}

// LinkWithID create a link component with a user specific id.
func LinkWithID(c *tgframe.Container, text, url, id string) {
	comp := newLinkComponent(text, url)
	comp.SetID(id)
	c.AddComponent(comp)
}
