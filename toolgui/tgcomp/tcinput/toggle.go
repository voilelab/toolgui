package tcinput

import (
	"github.com/voilelab/toolgui/toolgui/tgcomp/tcutil"
	"github.com/voilelab/toolgui/toolgui/tgframe"
)

var _ tgframe.Component = &toggleComponent{}
var toggleComponentName = "toggle_component"

type toggleComponent struct {
	*tgframe.BaseComponent
	Label    string `json:"label"`
	Default  bool   `json:"default"`
	Disabled bool   `json:"disabled"`
}

func newToggleComponent(label string) *toggleComponent {
	return &toggleComponent{
		BaseComponent: &tgframe.BaseComponent{
			Name: toggleComponentName,
			ID:   tcutil.NormalID(toggleComponentName, label),
		},
		Label: label,
	}
}

// ToggleConf is the configuration for a toggle.
type ToggleConf struct {
	tgframe.Base

	// Default is true if the toggle is default on.
	Default bool

	// Disabled is true if the toggle is disabled.
	Disabled bool
}

// Toggle create a switch and return true if it's on. It is [Checkbox] drawn as
// a switch: same signature, same value, different affordance.
func Toggle(c *tgframe.Container, label string, conf ...*ToggleConf) bool {
	cf := tgframe.OneConf("Toggle", conf)

	comp := newToggleComponent(label)
	comp.Default = cf.Default
	comp.Disabled = cf.Disabled
	tgframe.SetConfID(comp, cf)

	c.AddComponent(comp)
	return c.State.GetBool(comp.ID)
}
