package tcinput

import (
	"github.com/voilelab/toolgui/toolgui/tgframe"
)

var _ tgframe.Component = &formComponent{}
var formComponentName = "form_component"

type formComponent struct {
	*tgframe.BaseComponent
	SubmitLabel string `json:"submit_label"`
	HideSubmit  bool   `json:"hide_submit"`
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

	// SubmitLabel is the text on the built-in submit button. Empty is
	// "Submit".
	SubmitLabel string

	// HideSubmit drops the built-in submit button. Submitting is then up to a
	// [Button] written inside the form, which sends the form along with its
	// own click; a form with neither has no way to be sent.
	HideSubmit bool
}

// Form create a form component.
//
// The input components inside a form do not rerun the page as they are
// touched: their values are held on the client and sent together when the
// form is submitted, so the page reads a whole set of them at once.
//
// A form is submitted by its built-in submit button, or by a [Button] written
// inside it — a click inside a form carries the held values with it, so the
// page sees the new values on the run the button reports its click.
func Form(c *tgframe.Container, conf ...*FormConf) *tgframe.Container {
	cf := tgframe.OneConf("Form", conf)

	comp := newFormComponent()
	comp.SubmitLabel = cf.SubmitLabel
	comp.HideSubmit = cf.HideSubmit
	tgframe.SetConfID(comp, cf)

	formComp := c.AddComponent(comp)
	return c.AddContainerTo(formComp, "inner", 0)
}
