package tcinput

import (
	"slices"

	"github.com/voilelab/toolgui/toolgui/tgcomp/tcutil"
	"github.com/voilelab/toolgui/toolgui/tgframe"
)

var _ tgframe.Component = &multiselectComponent{}
var multiselectComponentName = "multiselect_component"

type multiselectComponent struct {
	*tgframe.BaseComponent
	Label         string   `json:"label"`
	Items         []string `json:"items"`
	Default       []int    `json:"default"`
	MaxSelections int      `json:"max_selections"`
	Placeholder   string   `json:"placeholder"`
	Disabled      bool     `json:"disabled"`
}

func newMultiselectComponent(label string, items []string) *multiselectComponent {
	return &multiselectComponent{
		BaseComponent: &tgframe.BaseComponent{
			Name: multiselectComponentName,
			ID:   tcutil.NormalID(multiselectComponentName, label),
		},
		Label: label,
		Items: items,
	}
}

// MultiselectConf is the configuration for the Multiselect component.
type MultiselectConf struct {
	tgframe.Base

	// Default is the selection the component starts with, as indices into
	// items. It is only read until the app user first touches the component.
	Default []int

	// MaxSelections is how many items may be selected at once. Zero, the
	// default, is no limit. At the limit the frontend disables the items that
	// are not selected, so the limit is never reached by a refusal.
	MaxSelections int

	// Placeholder is the text shown while nothing is selected.
	Placeholder string

	// Disabled is true if the multiselect is disabled.
	Disabled bool
}

// Multiselect create a dropdown list that takes more than one item and return
// the indices of the selected ones, 0-indexed.
//
// The result is ordered by items rather than by the order they were picked in,
// and is empty rather than nil when nothing is selected.
func Multiselect(c *tgframe.Container, label string, items []string,
	conf ...*MultiselectConf) []int {

	cf := tgframe.OneConf("Multiselect", conf)

	comp := newMultiselectComponent(label, items)
	comp.Default = normalizeSelection(cf.Default, len(items), cf.MaxSelections)
	comp.MaxSelections = cf.MaxSelections
	comp.Placeholder = cf.Placeholder
	comp.Disabled = cf.Disabled
	tgframe.SetConfID(comp, cf)

	c.AddComponent(comp)

	// The state holds whatever the select event landed, a []int here but a
	// []float64 once it has been through JSON; GetObject reads both, and
	// leaves idxes nil for a key holding neither.
	var idxes []int
	if c.State.GetObject(comp.ID, &idxes) != nil {
		idxes = nil
	}

	if idxes == nil {
		// Normalized a second time rather than handing back comp.Default: the
		// component is about to be serialized, and the caller owns what it
		// gets back.
		return normalizeSelection(cf.Default, len(items), cf.MaxSelections)
	}

	return normalizeSelection(idxes, len(items), cf.MaxSelections)
}

// normalizeSelection puts a selection into the shape the component promises:
// item order, no duplicates, nothing pointing outside items, and never nil.
// The cap is applied here too — the frontend is what keeps the app user from
// exceeding it, but a payload that did anyway is trimmed rather than trusted.
func normalizeSelection(idxes []int, itemCount, maxSelections int) []int {
	out := []int{}
	for _, idx := range idxes {
		if idx < 0 || idx >= itemCount {
			continue
		}

		out = append(out, idx)
	}

	slices.Sort(out)
	out = slices.Compact(out)

	if maxSelections > 0 && len(out) > maxSelections {
		out = out[:maxSelections]
	}

	return out
}
