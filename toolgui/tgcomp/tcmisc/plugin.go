package tcmisc

import (
	"github.com/voilelab/toolgui/toolgui/tgcomp/tcutil"
	"github.com/voilelab/toolgui/toolgui/tgframe"
)

var _ tgframe.Component = &pluginComponent{}
var pluginComponentName = "plugin_component"

const (
	// defaultPluginWidth is used when PluginConf.Width is empty.
	defaultPluginWidth = "100%"

	// defaultPluginHeight is used when PluginConf.Height is empty.
	defaultPluginHeight = "150px"
)

type pluginComponent struct {
	*tgframe.BaseComponent
	Src    string `json:"src"`
	Style  string `json:"style"`
	Props  any    `json:"props"`
	Width  string `json:"width"`
	Height string `json:"height"`
}

func newPluginComponent(src string) *pluginComponent {
	return &pluginComponent{
		BaseComponent: &tgframe.BaseComponent{
			Name: pluginComponentName,
			// The src stands in for an id the conf did not give, so a plugin
			// still claims a state key of its own and its value can be read
			// back. Two plugins running the same script need a conf id.
			ID: tcutil.HashedID(pluginComponentName, []byte(src)),
		},
		Src:    src,
		Width:  defaultPluginWidth,
		Height: defaultPluginHeight,
	}
}

// PluginConf is the configuration for the Plugin component.
type PluginConf struct {
	tgframe.Base

	// Props is handed to the plugin as its props, marshalled to json. The
	// plugin reads it through window.toolgui.onRender.
	Props any

	// Style is the url of a stylesheet to load in the frame, if the plugin
	// needs one. The frame gets none of the app's own css.
	Style string

	// Width is the css width of the plugin frame, default is "100%".
	Width string

	// Height is the css height of the plugin frame, default is "150px".
	// "auto" tracks the plugin's own height, which needs the plugin to call
	// window.toolgui.autoHeight().
	Height string
}

// Plugin runs the script at src in a sandboxed frame, and hands it props.
//
// The script is loaded from the app itself: register its files with
// [tgframe.App.AddPluginAssets] and name them with [tgframe.PluginAssetURL].
// It runs on an opaque origin, so it cannot reach the app's page, its dom, its
// cookies or its storage; it talks to the app through window.toolgui, the same
// bridge [Iframe] gives its html.
//
// Read what it sends back with [PluginValue].
func Plugin(c *tgframe.Container, src string, conf ...*PluginConf) {
	c.AddComponent(pluginComponentFor(src, tgframe.OneConf("Plugin", conf)))
}

// PluginValue returns the latest value the plugin src and conf describe sent
// through window.toolgui.update. It reads the value, it does not draw the
// plugin: give it the same src and conf the [Plugin] call gets, and read
// before drawing when the props depend on the value.
//
//	conf := &tcmisc.PluginConf{ID: "color_picker"}
//	selected := ""
//	if v := tcmisc.PluginValue[color](c, src, conf); v != nil {
//		selected = v.Color
//	}
//
//	conf.Props = map[string]any{"selected": selected}
//	tcmisc.Plugin(c, src, conf)
//
// It returns nil while the plugin has sent nothing, so a page tells "no value
// yet" apart from a value that is the zero T.
//
// The frontend keys the value by the plugin's own component id, so a plugin
// can only write to its own state.
func PluginValue[T any](c *tgframe.Container, src string, conf ...*PluginConf) *T {
	comp := pluginComponentFor(src, tgframe.OneConf("PluginValue", conf))

	return frameValue[T](c, "plugin", comp.ID)
}

// pluginComponentFor builds the component src and conf describe. Both entry
// points go through it, so the id PluginValue reads is the one Plugin draws.
func pluginComponentFor(src string, cf *PluginConf) *pluginComponent {
	comp := newPluginComponent(src)
	comp.Props = cf.Props
	comp.Style = cf.Style

	if cf.Width != "" {
		comp.Width = cf.Width
	}

	if cf.Height != "" {
		comp.Height = cf.Height
	}

	tgframe.SetConfID(comp, cf)

	return comp
}
