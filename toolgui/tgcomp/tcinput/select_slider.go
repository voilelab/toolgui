package tcinput

import (
	"fmt"

	"github.com/voilelab/toolgui/toolgui/tgcomp/tcutil"
	"github.com/voilelab/toolgui/toolgui/tgframe"
)

var _ tgframe.Component = &selectSliderComponent{}
var selectSliderComponentName = "select_slider_component"

type selectSliderComponent struct {
	*tgframe.BaseComponent

	Label    string   `json:"label"`
	Items    []string `json:"items"`
	Default  int      `json:"default"`
	Disabled bool     `json:"disabled"`
}

func newSelectSliderComponent(label string, items []string) *selectSliderComponent {
	return &selectSliderComponent{
		BaseComponent: &tgframe.BaseComponent{
			Name: selectSliderComponentName,
			ID:   tcutil.NormalID(selectSliderComponentName, label),
		},
		Label: label,
		Items: items,
	}
}

// SelectSliderConf is the configuration for the SelectSlider component.
type SelectSliderConf struct {
	tgframe.Base

	// Default is the index of the item the slider starts on. Defaults to 0.
	Default int

	// Disabled is the disabled state of the select slider component.
	Disabled bool
}

// SelectSlider create a slider over a list of discrete items and return the
// index of the one it sits on, 0-indexed as [Select]'s is.
//
// The handle is always on an item, so the return is never nil: it is the index
// the app user left it at, else Default. An empty items, or a Default outside
// it, is a mistake in the caller and panics.
func SelectSlider(c *tgframe.Container, label string, items []string,
	conf ...*SelectSliderConf) *int {

	cf := tgframe.OneConf("SelectSlider", conf)

	if len(items) == 0 {
		panic(fmt.Sprintf("toolgui: SelectSlider %q has no items", label))
	}

	if cf.Default < 0 || cf.Default >= len(items) {
		panic(fmt.Sprintf("toolgui: SelectSlider %q has default index %d "+
			"outside its %d items", label, cf.Default, len(items)))
	}

	comp := newSelectSliderComponent(label, items)
	comp.Default = cf.Default
	comp.Disabled = cf.Disabled
	tgframe.SetConfID(comp, cf)

	c.AddComponent(comp)

	// Unlike Select, which keeps 0 for "nothing selected", the index here is
	// 0-based on the wire too: there is no unselected position to reserve it
	// for. An index from an older run of a now shorter list is dropped rather
	// than clamped, so the value never disagrees with the handle.
	idx := c.State.GetInt(comp.ID)
	if idx == nil || *idx < 0 || *idx >= len(items) {
		return &comp.Default
	}

	return idx
}
