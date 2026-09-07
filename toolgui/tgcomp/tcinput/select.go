package tcinput

import (
	"github.com/voilelab/toolgui/toolgui/tgcomp/tcutil"
	"github.com/voilelab/toolgui/toolgui/tgframe"
)

var _ tgframe.Component = &selectComponent{}
var selectComponentName = "select_component"

type selectComponent struct {
	*tgframe.BaseComponent
	Label string   `json:"label"`
	Items []string `json:"items"`
}

func newSelectComponent(label string, items []string) *selectComponent {
	return &selectComponent{
		BaseComponent: &tgframe.BaseComponent{
			Name: selectComponentName,
			ID:   tcutil.NormalID(selectComponentName, label),
		},
		Label: label,
		Items: items,
	}
}

// SelectConf is the configuration for the Select component.
type SelectConf struct {
	// ID is the unique identifier for this select component.
	ID string
}

// Select create a select dropdown list and return its selected value.
// 0-indexed, return nil if no item is selected. conf may be nil.
func Select(c *tgframe.Container, label string, items []string, conf *SelectConf) *int {
	comp := newSelectComponent(label, items)
	if conf != nil && conf.ID != "" {
		comp.SetID(conf.ID)
	}

	c.AddComponent(comp)
	idx := c.State.GetInt(comp.ID)
	if idx == nil {
		return nil
	}

	if *idx == 0 {
		return nil
	}

	*idx--
	return idx
}
