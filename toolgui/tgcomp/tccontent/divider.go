package tccontent

import (
	"github.com/voilelab/toolgui/toolgui/tgcomp/tcutil"
	"github.com/voilelab/toolgui/toolgui/tgframe"
)

var _ tgframe.Component = &textComponent{}
var dividerComponentName = "divider_component"

type dividerComponent struct {
	*tgframe.BaseComponent
}

func newDividerComponent() *dividerComponent {
	return &dividerComponent{
		BaseComponent: &tgframe.BaseComponent{
			Name: dividerComponentName,
			ID:   tcutil.RandID(dividerComponentName),
		},
	}
}

// Divider create a horizontal line.
func Divider(c *tgframe.Container) {
	c.AddComponent(newDividerComponent())
}

// DividerWithID create a horizontal line with ID.
func DividerWithID(c *tgframe.Container, id string) {
	comp := newDividerComponent()
	comp.SetID(id)
	c.AddComponent(comp)
}
