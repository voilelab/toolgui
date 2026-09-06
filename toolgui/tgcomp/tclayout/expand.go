package tclayout

import (
	"github.com/voilelab/toolgui/toolgui/tgcomp/tcutil"
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
			// The client keeps whether this is open, so the component needs a
			// name of its own to keep that across runs.
			ID: tcutil.NormalID(expandComponentName, title),
		},

		Title:    title,
		Expanded: expanded,
	}
	return comp
}

// Expand create a expandable component.
func Expand(c *tgframe.Container, title string, expanded bool) *tgframe.Container {
	return expand(c, newExpandComponent(title, expanded))
}

// ExpandWithID create a expandable component with a user specific id.
func ExpandWithID(c *tgframe.Container, title string, expanded bool, id string) *tgframe.Container {
	expandComp := newExpandComponent(title, expanded)
	expandComp.SetID(id)
	return expand(c, expandComp)
}

func expand(c *tgframe.Container, expandComp *expandComponent) *tgframe.Container {
	comp := c.AddComponent(expandComp)
	return c.AddContainerTo(comp, "inner", 0)
}
