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
	// Default is true if the checkbox is default checked.
	Default bool

	// Disabled is true if the checkbox is disabled.
	Disabled bool

	// ID is the ID of the checkbox.
	ID string
}

// Checkbox create a checkbox and return true if it's clicked.
func Checkbox(s *tgframe.State, c *tgframe.Container, label string) bool {
	return CheckboxWithConf(s, c, label, nil)
}

// CheckboxWithConf create a checkbox and return true if it's clicked.
func CheckboxWithConf(s *tgframe.State, c *tgframe.Container, label string, conf *CheckboxConf) bool {
	if conf == nil {
		conf = &CheckboxConf{}
	}

	comp := newCheckboxComponent(label)
	if conf.ID != "" {
		comp.SetID(conf.ID)
	}

	comp.Default = conf.Default
	comp.Disabled = conf.Disabled

	c.AddComponent(comp)
	return s.GetBool(comp.ID)
}
