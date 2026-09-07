package tcinput

import (
	"github.com/voilelab/toolgui/toolgui/tgcomp/tcutil"
	"github.com/voilelab/toolgui/toolgui/tgframe"
)

var _ tgframe.Component = &radioComponent{}
var radioComponentName = "radio_component"

type radioComponent struct {
	*tgframe.BaseComponent
	Label string   `json:"label"`
	Items []string `json:"items"`
}

func newRadioComponent(label string, items []string) *radioComponent {
	return &radioComponent{
		BaseComponent: &tgframe.BaseComponent{
			Name: radioComponentName,
			ID:   tcutil.NormalID(radioComponentName, label),
		},
		Label: label,
		Items: items,
	}
}

// RadioConf is the configuration for the Radio component.
type RadioConf struct {
	tgframe.Base
}

// Radio create a group of radio items and return its selected value.
func Radio(c *tgframe.Container, label string, items []string, conf ...*RadioConf) *int {
	cf := tgframe.OneConf("Radio", conf)

	comp := newRadioComponent(label, items)
	tgframe.SetConfID(comp, cf)
	c.AddComponent(comp)

	idx := c.State.GetInt(comp.ID)
	if idx == nil {
		return nil
	}

	return idx
}
