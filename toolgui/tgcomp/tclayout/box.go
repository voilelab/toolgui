package tclayout

import (
	"github.com/voilelab/toolgui/toolgui/tgcomp/tcutil"
	"github.com/voilelab/toolgui/toolgui/tgframe"
)

var _ tgframe.Component = &boxComponent{}
var boxComponentName = "box_component"

type boxComponent struct {
	*tgframe.BaseComponent
}

func newBoxComponent(id string) *boxComponent {
	return &boxComponent{
		BaseComponent: &tgframe.BaseComponent{
			Name: boxComponentName,
			ID:   tcutil.NormalID(boxComponentName, id),
		},
	}
}

// Box create a box container.
func Box(c *tgframe.Container, id string) *tgframe.Container {
	boxComp := newBoxComponent(id)
	c.AddComponent(boxComp)

	cont := tgframe.NewContainer(boxComp.ID+"_inner", c.SendNotifyPack)
	c.SendNotifyPack(tgframe.NewNotifyPackCreate(boxComp.ID, cont))

	return cont
}
