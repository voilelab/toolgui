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
	tgframe.Base
}

// Select create a select dropdown list and return its selected value.
// 0-indexed, return nil if no item is selected.
func Select(c *tgframe.Container, label string, items []string, conf ...*SelectConf) *int {
	cf := tgframe.OneConf("Select", conf)

	comp := newSelectComponent(label, items)
	tgframe.SetConfID(comp, cf)

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
