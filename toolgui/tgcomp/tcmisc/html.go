package tcmisc

import (
	"github.com/voilelab/toolgui/toolgui/tgframe"
)

var _ tgframe.Component = &htmlComponent{}
var htmlComponentName = "html_component"

type htmlComponent struct {
	*tgframe.BaseComponent
	HTML string `json:"html"`
}

func newHTMLComponent(html string) *htmlComponent {
	return &htmlComponent{
		BaseComponent: &tgframe.BaseComponent{
			Name: htmlComponentName,
		},
		HTML: html,
	}
}

// HTMLConf is the configuration for the HTML component.
type HTMLConf struct {
	tgframe.Base
}

// HTML adds an HTML component to the container.
func HTML(c *tgframe.Container, html string, conf ...*HTMLConf) {
	cf := tgframe.OneConf("HTML", conf)

	comp := newHTMLComponent(html)
	tgframe.SetConfID(comp, cf)
	c.AddComponent(comp)
}
