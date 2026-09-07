package tccontent

import (
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
		},
	}
}

// DividerConf is the configuration for the Divider component.
type DividerConf struct {
	tgframe.Base
}

// Divider create a horizontal line.
func Divider(c *tgframe.Container, conf ...*DividerConf) {
	cf := tgframe.OneConf("Divider", conf)

	comp := newDividerComponent()
	tgframe.SetConfID(comp, cf)
	c.AddComponent(comp)
}
