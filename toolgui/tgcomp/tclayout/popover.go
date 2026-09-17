package tclayout

import (
	"github.com/voilelab/toolgui/toolgui/tgcomp/tcutil"
	"github.com/voilelab/toolgui/toolgui/tgframe"
)

var _ tgframe.Component = &popoverComponent{}

const popoverComponentName = "popover_component"

type popoverComponent struct {
	*tgframe.BaseComponent

	Label    string `json:"label"`
	Disabled bool   `json:"disabled"`
}

func newPopoverComponent(label string) *popoverComponent {
	return &popoverComponent{
		BaseComponent: &tgframe.BaseComponent{
			Name: popoverComponentName,
			// The client keeps whether this is open, so the component needs a
			// name of its own to keep that across runs.
			ID: tcutil.NormalID(popoverComponentName, label),
		},

		Label: label,
	}
}

// PopoverConf is the configuration for the Popover component.
type PopoverConf struct {
	tgframe.Base

	// Disabled is true if the trigger button is disabled.
	Disabled bool
}

// Popover create a button with a floating panel behind it, and return the
// container the panel holds.
func Popover(c *tgframe.Container, label string, conf ...*PopoverConf) *tgframe.Container {
	cf := tgframe.OneConf("Popover", conf)

	popoverComp := newPopoverComponent(label)
	popoverComp.Disabled = cf.Disabled
	tgframe.SetConfID(popoverComp, cf)

	comp := c.AddComponent(popoverComp)
	return c.AddContainerTo(comp, "inner", 0)
}
