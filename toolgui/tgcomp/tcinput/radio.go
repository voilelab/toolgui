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

	// Default is 0-based, as the state is. No index is left over to mean "no
	// default", so it is a pointer and null is one.
	Default  *int `json:"default"`
	Disabled bool `json:"disabled"`
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

	// Default is the item the group starts on, as an index into items,
	// 0-based like the return. It is only read until the app user first
	// touches the component, and one that points outside items is ignored.
	Default *int

	// Disabled is true if the radio group is disabled.
	Disabled bool
}

// SetDefault sets Default from a value, so a literal index can be written
// where the conf is.
func (c *RadioConf) SetDefault(v int) *RadioConf {
	c.Default = &v
	return c
}

// Radio create a group of radio items and return its selected value.
// 0-indexed, return nil if no item is selected.
//
// A radio has no placeholder to reserve an index for, so the state under its
// id is 0-based, unlike [Select]'s. Neither shows through here.
func Radio(c *tgframe.Container, label string, items []string, conf ...*RadioConf) *int {
	cf := tgframe.OneConf("Radio", conf)

	comp := newRadioComponent(label, items)
	comp.Default = normalizeIndex(cf.Default, len(items))
	comp.Disabled = cf.Disabled
	tgframe.SetConfID(comp, cf)
	c.AddComponent(comp)

	idx := c.State.GetInt(comp.ID)
	if idx == nil {
		// Untouched, so the default stands in — as its own pointer, which
		// the caller then owns.
		return normalizeIndex(cf.Default, len(items))
	}

	// As in [Select], a selection left over from a longer list is dropped
	// rather than clamped.
	return normalizeIndex(idx, len(items))
}
