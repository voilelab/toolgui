package tcmisc

import (
	"github.com/voilelab/toolgui/toolgui/tgcomp/tcutil"
	"github.com/voilelab/toolgui/toolgui/tgframe"
)

var _ tgframe.Component = &iframeComponent{}
var iframeComponentName = "iframe_component"

type iframeComponent struct {
	*tgframe.BaseComponent
	Html   string `json:"html"`
	Script bool   `json:"script"`
}

func newIframeComponent(html string, script bool) *iframeComponent {
	return &iframeComponent{
		BaseComponent: &tgframe.BaseComponent{
			Name: iframeComponentName,
			ID:   tcutil.HashedID(iframeComponentName, []byte(html)),
		},
		Html:   html,
		Script: script,
	}
}

// Iframe show a html.
// script is used to allow the iframe to run javascript. (notice that this is not secure)
func Iframe(c *tgframe.Container, html string, script bool) {
	comp := newIframeComponent(html, script)
	c.AddComponent(comp)
}

// IframeWithID create a html component with a user specific id.
// script is used to allow the iframe to run javascript. (notice that this is not secure)
func IframeWithID(c *tgframe.Container, html string, script bool, id string) {
	comp := newIframeComponent(html, script)
	comp.SetID(id)
	c.AddComponent(comp)
}
