package tclayout

import (
	"github.com/voilelab/toolgui/toolgui/tgframe"
)

var _ tgframe.Component = &expandComponent{}

const expandComponentName = "expand_component"

type expandComponent struct {
	*tgframe.BaseComponent

	Title    string `json:"title"`
	Expanded bool   `json:"expanded"`
}

func newExpandComponent(title string, expanded bool) *expandComponent {
	comp := &expandComponent{
		BaseComponent: &tgframe.BaseComponent{
			Name: expandComponentName,
		},

		Title:    title,
		Expanded: expanded,
	}
	return comp
}

// Expand create a expandable component.
func Expand(c *tgframe.Container, title string, expanded bool) *tgframe.Container {
	comp := c.AddComponent(newExpandComponent(title, expanded))
	return c.AddContainerTo(comp, "inner", 0)
}
