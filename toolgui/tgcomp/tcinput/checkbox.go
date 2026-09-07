package tcinput

import (
	"github.com/voilelab/toolgui/toolgui/tgcomp/tcutil"
	"github.com/voilelab/toolgui/toolgui/tgframe"
)

var _ tgframe.Component = &checkboxComponent{}
var checkboxComponentName = "checkbox_component"

type checkboxComponent struct {
	*tgframe.BaseComponent
	Label    string `json:"label"`
	Default  bool   `json:"default"`
	Disabled bool   `json:"disabled"`
}

func newCheckboxComponent(label string) *checkboxComponent {
	return &checkboxComponent{
		BaseComponent: &tgframe.BaseComponent{
			Name: checkboxComponentName,
			ID:   tcutil.NormalID(checkboxComponentName, label),
		},
		Label: label,
	}
}

// CheckboxConf is the configuration for a checkbox.
type CheckboxConf struct {
	tgframe.Base

	// Default is true if the checkbox is default checked.
	Default bool

	// Disabled is true if the checkbox is disabled.
	Disabled bool
}

// Checkbox create a checkbox and return true if it's checked.
func Checkbox(c *tgframe.Container, label string, conf ...*CheckboxConf) bool {
	cf := tgframe.OneConf("Checkbox", conf)

	comp := newCheckboxComponent(label)
	comp.Default = cf.Default
	comp.Disabled = cf.Disabled
	tgframe.SetConfID(comp, cf)

	c.AddComponent(comp)
	return c.State.GetBool(comp.ID)
}
