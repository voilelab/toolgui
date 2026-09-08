package tcinput

import (
	"fmt"
	"strings"

	"github.com/voilelab/toolgui/toolgui/tgcomp/tcutil"
	"github.com/voilelab/toolgui/toolgui/tgframe"
)

var _ tgframe.Component = &colorPickerComponent{}
var colorPickerComponentName = "color_picker_component"

// The color a picker shows before anything is picked, and what an unset
// Default means.
const defaultPickerColor = "#000000"

type colorPickerComponent struct {
	*tgframe.BaseComponent
	Label    string `json:"label"`
	Default  string `json:"default"`
	Disabled bool   `json:"disabled"`
}

func newColorPickerComponent(label string) *colorPickerComponent {
	return &colorPickerComponent{
		BaseComponent: &tgframe.BaseComponent{
			Name: colorPickerComponentName,
			ID:   tcutil.NormalID(colorPickerComponentName, label),
		},
		Label: label,
	}
}

// ColorPickerConf is the configuration for a color picker.
type ColorPickerConf struct {
	tgframe.Base

	// Default is the color the picker starts on, as "#rrggbb". Defaults to
	// black.
	Default string

	// Disabled is true if the color picker is disabled.
	Disabled bool
}

// ColorPicker create a color picker and return the picked color as a lowercase
// "#rrggbb" string. Before anything is picked that is the conf's Default.
//
// A Default that is not an "#rrggbb" color is a mistake in the caller rather
// than a value to correct, and panics.
func ColorPicker(c *tgframe.Container, label string, conf ...*ColorPickerConf) string {
	cf := tgframe.OneConf("ColorPicker", conf)

	def := cf.Default
	if def == "" {
		def = defaultPickerColor
	}

	if !isHexColor(def) {
		panic(fmt.Sprintf("toolgui: ColorPicker %q has default %q, "+
			"which is not an #rrggbb color", label, cf.Default))
	}

	comp := newColorPickerComponent(label)
	comp.Default = strings.ToLower(def)
	comp.Disabled = cf.Disabled
	tgframe.SetConfID(comp, cf)

	c.AddComponent(comp)

	// The client is asked for "#rrggbb", but the state can also have been
	// written from Go, so anything that is not one reads as unpicked rather
	// than reaching the page as a color it cannot use.
	val := c.State.GetString(comp.ID)
	if val == nil || !isHexColor(*val) {
		return comp.Default
	}

	return strings.ToLower(*val)
}

// isHexColor reports whether s is an "#rrggbb" color, in either case.
func isHexColor(s string) bool {
	if len(s) != 7 || s[0] != '#' {
		return false
	}

	for _, r := range s[1:] {
		switch {
		case r >= '0' && r <= '9':
		case r >= 'a' && r <= 'f':
		case r >= 'A' && r <= 'F':
		default:
			return false
		}
	}

	return true
}
