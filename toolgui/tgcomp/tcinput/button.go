package tcinput

import (
	"github.com/voilelab/toolgui/toolgui/tgcomp/tcutil"
	"github.com/voilelab/toolgui/toolgui/tgframe"
)

var _ tgframe.Component = &buttonComponent{}
var buttonComponentName = "button_component"

type buttonComponent struct {
	*tgframe.BaseComponent
	Label    string       `json:"label"`
	Color    tcutil.Color `json:"color"`
	Disabled bool         `json:"disabled"`
}

func newButtonComponent(label string) *buttonComponent {
	return &buttonComponent{
		BaseComponent: &tgframe.BaseComponent{
			Name: buttonComponentName,
			ID:   tcutil.NormalID(buttonComponentName, label),
		},
		Label: label,
	}
}

// ButtonConf is the configuration for the Button component
type ButtonConf struct {
	tgframe.Base

	// Color defines the color of the button
	Color tcutil.Color

	// Disabled indicates whether the button should be initially disabled
	Disabled bool
}

// Button create a button and return true if it's clicked.
func Button(c *tgframe.Container, label string, conf ...*ButtonConf) bool {
	cf := tgframe.OneConf("Button", conf)

	comp := newButtonComponent(label)
	comp.Color = cf.Color
	comp.Disabled = cf.Disabled
	tgframe.SetConfID(comp, cf)

	c.AddComponent(comp)
	return c.State.GetClickID() == comp.ID
}
