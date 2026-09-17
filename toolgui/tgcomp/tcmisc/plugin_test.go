package tcmisc

import (
	"testing"

	"github.com/voilelab/toolgui/toolgui/tgcomp/tcutil"
	"github.com/voilelab/toolgui/toolgui/tgframe"
)

func TestPluginProps(t *testing.T) {
	props := addIframe(t, func(c *tgframe.Container) {
		Plugin(c, "/plugin/gauge/gauge.js", &PluginConf{
			ID:    "my_plugin",
			Props: map[string]any{"value": 42},
		})
	})

	if props["src"] != "/plugin/gauge/gauge.js" {
		t.Errorf("src = %v", props["src"])
	}

	if props["id"] != "plugin_component_my_plugin" {
		t.Errorf("id = %v, want plugin_component_my_plugin", props["id"])
	}

	if props["width"] != defaultPluginWidth {
		t.Errorf("width = %v, want %v", props["width"], defaultPluginWidth)
	}

	if props["height"] != defaultPluginHeight {
		t.Errorf("height = %v, want %v", props["height"], defaultPluginHeight)
	}

	sent, ok := props["props"].(map[string]any)
	if !ok {
		t.Fatalf("props = %T, want an object", props["props"])
	}

	if sent["value"] != float64(42) {
		t.Errorf("props.value = %v, want 42", sent["value"])
	}
}

func TestPluginConfDrivesTheProps(t *testing.T) {
	props := addIframe(t, func(c *tgframe.Container) {
		Plugin(c, "/plugin/gauge/gauge.js", &PluginConf{
			ID:     "my_plugin",
			Style:  "/plugin/gauge/gauge.css",
			Width:  "300px",
			Height: "auto",
		})
	})

	if props["style"] != "/plugin/gauge/gauge.css" {
		t.Errorf("style = %v", props["style"])
	}

	if props["width"] != "300px" {
		t.Errorf("width = %v, want 300px", props["width"])
	}

	if props["height"] != "auto" {
		t.Errorf("height = %v, want auto", props["height"])
	}
}

// A plugin with no conf id derives one from its src, so it still claims a
// state key and its value can be read back.
func TestPluginWithoutID(t *testing.T) {
	const src = "/plugin/gauge/gauge.js"

	props := addIframe(t, func(c *tgframe.Container) {
		Plugin(c, src)
	})

	want := tcutil.HashedID(pluginComponentName, []byte(src))
	if props["id"] != want {
		t.Errorf("id = %v, want %v", props["id"], want)
	}
}

// A value the plugin sent comes back typed, keyed by the plugin's own
// component id rather than by the bare user id.
func TestPluginValueRoundTrip(t *testing.T) {
	const src = "/plugin/gauge/gauge.js"
	conf := &PluginConf{ID: "my_plugin"}

	type gauge struct {
		Value int `json:"value"`
	}

	state := stateWithValue(t, map[string]any{"value": 7}, func(c *tgframe.Container) {
		Plugin(c, src, conf)
	})

	var got *gauge
	drawWithState(t, state, func(c *tgframe.Container) {
		got = PluginValue[gauge](c, src, conf)
	})

	if got == nil {
		t.Fatal("value = nil, want the value the plugin sent")
	}

	if got.Value != 7 {
		t.Errorf("value = %d, want 7", got.Value)
	}
}

// Before the plugin sends anything there is no value, which is not the same as
// a value that is the zero T.
func TestPluginValueBeforeTheFirstUpdate(t *testing.T) {
	type gauge struct {
		Value int `json:"value"`
	}

	var got *gauge
	drawWithState(t, tgframe.NewState(), func(c *tgframe.Container) {
		got = PluginValue[gauge](c, "/plugin/gauge/gauge.js")
	})

	if got != nil {
		t.Errorf("value = %v, want nil before the plugin sends one", got)
	}
}

// A value that does not fit T is the run's failure, not a silent zero.
func TestPluginValueThatDoesNotParse(t *testing.T) {
	const src = "/plugin/gauge/gauge.js"

	type gauge struct {
		Value int `json:"value"`
	}

	state := stateWithValue(t, map[string]any{"value": "not a number"},
		func(c *tgframe.Container) {
			Plugin(c, src)
		})

	var got *gauge
	names := drawWithState(t, state, func(c *tgframe.Container) {
		got = PluginValue[gauge](c, src)
	})

	if got != nil {
		t.Errorf("value = %v, want nil when the value does not parse", got)
	}

	// The failure is on screen, not only in the log.
	if len(names) != 1 || names[0] != tgframe.ErrorComponentName {
		t.Errorf("drew %v, want an error placeholder", names)
	}
}
