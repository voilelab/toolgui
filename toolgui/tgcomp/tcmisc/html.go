package tcmisc

import (
	"github.com/voilelab/toolgui/toolgui/tgframe"
)

var _ tgframe.Component = &htmlComponent{}
var htmlComponentName = "html_component"

type htmlComponent struct {
	*tgframe.BaseComponent
	Html string `json:"html"`
}

func newHtmlComponent(html string) *htmlComponent {
	return &htmlComponent{
		BaseComponent: &tgframe.BaseComponent{
			Name: htmlComponentName,
		},
		Html: html,
	}
}

// HtmlConf is the configuration for the Html component.
type HtmlConf struct {
	tgframe.Base
}

// Html adds an HTML component to the container.
func Html(c *tgframe.Container, html string, conf ...*HtmlConf) {
	cf := tgframe.OneConf("Html", conf)

	comp := newHtmlComponent(html)
	tgframe.SetConfID(comp, cf)
	c.AddComponent(comp)
}
