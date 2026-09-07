package tccontent

import (
	"github.com/voilelab/toolgui/toolgui/tgcomp/tcutil"
	"github.com/voilelab/toolgui/toolgui/tgframe"
)

var _ tgframe.Component = &linkButtonComponent{}
var linkButtonComponentName = "link_button_component"

type linkButtonComponent struct {
	*tgframe.BaseComponent
	Text  string       `json:"text"`
	URL   string       `json:"url"`
	Color tcutil.Color `json:"color"`
}

func newLinkButtonComponent(text, url string) *linkButtonComponent {
	return &linkButtonComponent{
		BaseComponent: &tgframe.BaseComponent{
			Name: linkButtonComponentName,
		},
		Text: text,
		URL:  url,
	}
}

// LinkButtonConf is the configuration for the LinkButton component.
type LinkButtonConf struct {
	tgframe.Base

	// Color is the color of the button. Default is tcutil.ColorNull, which
	// leaves it neutral.
	Color tcutil.Color
}

// LinkButton create a link that is drawn as a button. It navigates rather
// than reporting a click, so unlike Button it returns nothing and keeps no
// state.
func LinkButton(c *tgframe.Container, text, url string, conf ...*LinkButtonConf) {
	cf := tgframe.OneConf(conf)

	comp := newLinkButtonComponent(text, url)
	comp.Color = cf.Color
	tgframe.SetConfID(comp, cf)
	c.AddComponent(comp)
}
