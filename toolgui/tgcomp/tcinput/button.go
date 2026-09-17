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
	comp := buttonComponentFor(label, tgframe.OneConf("Button", conf))

	c.AddComponent(comp)
	return c.State.GetClickID() == comp.ID
}

// ButtonClicked reports whether the click this run is handling is the one on
// the button label and conf describe. It reads the click, it does not draw the
// button: give it the same label and conf the [Button] call gets.
//
// It reads the click off the run's state rather than off the button, so a page
// can ask before the button is drawn — the case for a button written below the
// content it changes, which would otherwise have to send the old content out
// first and rewrite it.
//
//	if tcinput.ButtonClicked(c, "Load details") {
//		details = load()
//	}
//	...
//	tcinput.Button(c, "Load details")
//
// The click id comes from the client, and the page has not drawn anything yet
// to check it against, so the button the last run put on the screen is what
// it's checked against — the same guard an upload naming a component id goes
// through. A click on a button that was not there is not a click.
func ButtonClicked(c *tgframe.Container, label string, conf ...*ButtonConf) bool {
	comp := buttonComponentFor(label, tgframe.OneConf("ButtonClicked", conf))

	return clicked(c, comp.ID)
}

// buttonComponentFor builds the component label and conf describe. Both entry
// points go through it, so the id ButtonClicked reads is the one Button draws.
func buttonComponentFor(label string, cf *ButtonConf) *buttonComponent {
	comp := newButtonComponent(label)
	comp.Color = cf.Color
	comp.Disabled = cf.Disabled
	tgframe.SetConfID(comp, cf)

	return comp
}

// clicked reports whether the click this run is handling is the one on the
// component with the given component id, and that the last run drew it.
func clicked(c *tgframe.Container, compID string) bool {
	return c.State.GetClickID() == compID && c.State.HasComponentID(compID)
}
