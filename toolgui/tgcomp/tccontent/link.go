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

// LinkConf is the configuration for the Link component.
type LinkConf struct {
	tgframe.Base
}

// Link create a link component.
func Link(c *tgframe.Container, text, url string, conf ...*LinkConf) {
	cf := tgframe.OneConf("Link", conf)

	comp := newLinkComponent(text, url)
	tgframe.SetConfID(comp, cf)
	c.AddComponent(comp)
}
