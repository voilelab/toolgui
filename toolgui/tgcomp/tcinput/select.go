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

	// Default is 1-based like the state, 0 being "no default", so it matches
	// the option values the frontend builds.
	Default  int  `json:"default"`
	Disabled bool `json:"disabled"`
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

	// Default is the item the select starts on, as an index into items,
	// 0-based like the return. It is only read until the app user first
	// touches the component, and one that points outside items is ignored.
	Default *int

	// Disabled is true if the select is disabled.
	Disabled bool
}

// SetDefault sets Default from a value, so a literal index can be written
// where the conf is.
func (c *SelectConf) SetDefault(v int) *SelectConf {
	c.Default = &v
	return c
}

// Select create a select dropdown list and return its selected value.
// 0-indexed, return nil if no item is selected.
//
// The state under the component's id is 1-based, keeping 0 for "nothing
// selected" because the frontend's first option is a placeholder. That offset
// is entirely internal: both Conf.Default and the return are 0-based, so
// Default: 0 is items[0].
func Select(c *tgframe.Container, label string, items []string, conf ...*SelectConf) *int {
	cf := tgframe.OneConf("Select", conf)

	comp := newSelectComponent(label, items)
	if def := normalizeIndex(cf.Default, len(items)); def != nil {
		comp.Default = *def + 1
	}
	comp.Disabled = cf.Disabled
	tgframe.SetConfID(comp, cf)

	c.AddComponent(comp)
	idx := c.State.GetInt(comp.ID)
	if idx == nil {
		// Untouched, so the default stands in — as its own pointer, which
		// the caller then owns.
		return normalizeIndex(cf.Default, len(items))
	}

	if *idx == 0 {
		return nil
	}

	// A selection left over from a longer list is dropped rather than
	// clamped, the way [Multiselect] drops one.
	*idx--
	return normalizeIndex(idx, len(items))
}

// normalizeIndex copies an index that points at an item and drops one that
// does not, so a Default outside items reads as no default rather than taking
// the page down. [Radio] uses it too.
func normalizeIndex(idx *int, itemCount int) *int {
	if idx == nil || *idx < 0 || *idx >= itemCount {
		return nil
	}

	v := *idx
	return &v
}
