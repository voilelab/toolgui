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

// ExpandConf is the configuration for the Expand component.
type ExpandConf struct {
	tgframe.Base
}

// Expand create a expandable component.
func Expand(c *tgframe.Container, title string, expanded bool, conf ...*ExpandConf) *tgframe.Container {
	cf := tgframe.OneConf("Expand", conf)

	expandComp := newExpandComponent(title, expanded)
	tgframe.SetConfID(expandComp, cf)

	comp := c.AddComponent(expandComp)
	return c.AddContainerTo(comp, "inner", 0)
}
