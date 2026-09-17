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

// ButtonClicked reports whether the click this run is handling is the one on
// the button with the given id. The id is the one passed as [ButtonConf.ID].
//
// It reads the click off the state rather than off the button, so a page can
// ask before the button is drawn — the case for a button written below the
// content it changes, which would otherwise have to send the old content out
// first and rewrite it.
//
// Comparing an id with [tgframe.State.GetClickID] directly does not work:
// GetClickID returns the button's component id, which carries the component
// name in front of the conf id.
func ButtonClicked(s *tgframe.State, id string) bool {
	return s.GetClickID() == tcutil.NormalID(buttonComponentName, id)
}
