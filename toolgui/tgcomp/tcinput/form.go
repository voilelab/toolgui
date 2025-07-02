package tcinput

import (
	"github.com/voilelab/toolgui/toolgui/tgcomp/tcutil"
	"github.com/voilelab/toolgui/toolgui/tgframe"
)

var _ tgframe.Component = &formComponent{}
var formComponentName = "form_component"

type formComponent struct {
	*tgframe.BaseComponent
}

func newFormComponent(id string) *formComponent {
	return &formComponent{
		BaseComponent: &tgframe.BaseComponent{
			Name: formComponentName,
			ID:   tcutil.NormalID(formComponentName, id),
		},
	}
}

// Form create a form component.
func Form(c *tgframe.Container, id string) *tgframe.Container {
	formComp := newFormComponent(id)
	c.AddComponent(formComp)

	cont := tgframe.NewContainer(formComp.ID+"_inner", c.SendNotifyPack)
	c.SendNotifyPack(tgframe.NewNotifyPackCreate(formComp.ID, cont))

	return cont
}
