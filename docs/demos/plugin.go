package demos

import (
	"github.com/voilelab/toolgui/toolgui/tgcomp"
	"github.com/voilelab/toolgui/toolgui/tgframe"
)

// pickedColor is what the colorpicker plugin sends through
// window.toolgui.update.
//
// ANCHOR: value
type pickedColor struct {
	Color string `json:"color"`
}

// ANCHOR_END: value

// PickerColors are the swatches the plugin demo offers. The demo app hands
// the plugin's own files to the executor, so the list lives here beside the
// example that reads it.
var PickerColors = []string{"#ff3860", "#ffdd57", "#23d160", "#3273dc", "#b86bff"}

func pluginDemo(p *tgframe.Params) error {
	// ANCHOR: demo
	src := tgframe.PluginAssetURL("colorpicker", "colorpicker.js")
	conf := &tgcomp.PluginConf{
		ID:     "color_picker",
		Style:  tgframe.PluginAssetURL("colorpicker", "colorpicker.css"),
		Height: "auto",
	}

	// The plugin keeps no state of its own, so what is selected has to be
	// read before it is drawn. Nothing is until it sends its first value.
	selected := ""
	if v := tgcomp.PluginValue[pickedColor](p.Main, src, conf); v != nil {
		selected = v.Color
	}

	conf.Props = map[string]any{
		"colors":   PickerColors,
		"selected": selected,
	}

	tgcomp.Plugin(p.Main, src, conf)

	tgcomp.Text(p.Main, "Selected: "+selected)
	// ANCHOR_END: demo
	return nil
}
