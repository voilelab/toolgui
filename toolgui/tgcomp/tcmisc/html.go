package tcmisc

import (
	"github.com/voilelab/toolgui/toolgui/tgcomp/tcutil"
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
			ID:   tcutil.HashedID(htmlComponentName, []byte(html)),
		},
		Html: html,
	}
}

// Html adds a html component to the container.
func Html(c *tgframe.Container, html string) {
	comp := newHtmlComponent(html)
	c.AddComponent(comp)
}

// HtmlWithID adds a html component to the container with a specific id.
func HtmlWithID(c *tgframe.Container, html string, id string) {
	comp := newHtmlComponent(html)
	comp.SetID(id)
	c.AddComponent(comp)
}
