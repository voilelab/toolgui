package tcinput

import (
	"github.com/voilelab/toolgui/toolgui/tgframe"
)

var _ tgframe.Component = &formComponent{}
var formComponentName = "form_component"

type formComponent struct {
	*tgframe.BaseComponent
}

func newFormComponent() *formComponent {
	return &formComponent{
		BaseComponent: &tgframe.BaseComponent{
			Name: formComponentName,
		},
	}
}

// FormConf is the configuration for the Form component. The container a form
// hands out derives its id from the form's; give none and it carries none, and
// the components inside are still placed by position.
type FormConf struct {
	tgframe.Base
}

// Form create a form component.
func Form(c *tgframe.Container, conf ...*FormConf) *tgframe.Container {
	cf := tgframe.OneConf("Form", conf)

	comp := newFormComponent()
	tgframe.SetConfID(comp, cf)

	formComp := c.AddComponent(comp)
	return c.AddContainerTo(formComp, "inner", 0)
}
