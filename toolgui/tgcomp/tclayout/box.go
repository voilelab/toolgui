package tclayout

import (
	"github.com/voilelab/toolgui/toolgui/tgframe"
)

var _ tgframe.Component = &boxComponent{}
var boxComponentName = "box_component"

type boxComponent struct {
	*tgframe.BaseComponent
}

func newBoxComponent() *boxComponent {
	return &boxComponent{
		BaseComponent: &tgframe.BaseComponent{
			Name: boxComponentName,
		},
	}
}

// BoxConf is the configuration for the Box component. The container a box
// hands out derives its id from the box's; give none and it carries none, and
// the components inside are still placed by position.
type BoxConf struct {
	tgframe.Base
}

// Box create a box container.
func Box(c *tgframe.Container, conf ...*BoxConf) *tgframe.Container {
	cf := tgframe.OneConf("Box", conf)

	comp := newBoxComponent()
	tgframe.SetConfID(comp, cf)

	boxComp := c.AddComponent(comp)
	return c.AddContainerTo(boxComp, "inner", 0)
}
