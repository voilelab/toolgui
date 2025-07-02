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

// Select create a select dropdown list and return its selected value.
// 0-indexed, return nil if no item is selected.
func Select(s *tgframe.State, c *tgframe.Container, label string, items []string) *int {
	comp := newSelectComponent(label, items)
	c.AddComponent(comp)
	idx := s.GetInt(comp.ID)
	if idx == nil {
		return nil
	}

	if *idx == 0 {
		return nil
	}

	*idx--
	return idx
}
