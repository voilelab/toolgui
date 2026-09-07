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

func newPluginComponent(id, src string) *pluginComponent {
	comp := &pluginComponent{
		BaseComponent: &tgframe.BaseComponent{
			Name: pluginComponentName,
		},
		Src:    src,
		Width:  defaultPluginWidth,
		Height: defaultPluginHeight,
	}

	if id != "" {
		comp.SetID(id)
	}

	return comp
}

// PluginConf is the configuration for the Plugin component.
type PluginConf struct {
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
// The id names the plugin's state: it is the key [PluginValue] reads and the
// id the frontend stamps on every value the plugin sends. A plugin with an
// empty id renders, but cannot send anything back.
func Plugin(c *tgframe.Container, id, src string, props any) {
	PluginWithConf(c, id, src, &PluginConf{Props: props})
}

// PluginWithConf runs a plugin with a custom configuration.
func PluginWithConf(c *tgframe.Container, id, src string, conf *PluginConf) {
	if conf == nil {
		conf = &PluginConf{}
	}

	comp := newPluginComponent(id, src)
	comp.Props = conf.Props
	comp.Style = conf.Style

	if conf.Width != "" {
		comp.Width = conf.Width
	}

	if conf.Height != "" {
		comp.Height = conf.Height
	}

	c.AddComponent(comp)
}

// PluginValue unmarshals the latest value the plugin with the given id sent
// through window.toolgui.update into out. The id is the one passed to [Plugin].
//
// The frontend keys the value by the plugin's own component id, so a plugin
// can only write to its own state.
func PluginValue(s *tgframe.State, id string, out any) error {
	return s.GetObject(tcutil.NormalID(pluginComponentName, id), out)
}
