package tcmisc

import (
	"testing"

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

// A plugin with no id claims no state key, so nothing collides and nothing
// can be read back.
func TestPluginWithoutID(t *testing.T) {
	props := addIframe(t, func(c *tgframe.Container) {
		Plugin(c, "/plugin/gauge/gauge.js")
	})

	if props["id"] != "" {
		t.Errorf("id = %v, want empty", props["id"])
	}
}

// The state key PluginValue reads must be the component id the frontend puts
// on the event, not the bare user id.
func TestPluginValueRoundTrip(t *testing.T) {
	props := addIframe(t, func(c *tgframe.Container) {
		Plugin(c, "/plugin/gauge/gauge.js", &PluginConf{ID: "my_plugin"})
	})

	event := &tgframe.EventCustom{
		ID:    props["id"].(string),
		Value: map[string]any{"value": 7},
	}

	state := tgframe.NewState()
	event.ApplyState(state)

	var out struct {
		Value int `json:"value"`
	}
	if err := PluginValue(state, "my_plugin", &out); err != nil {
		t.Fatalf("PluginValue: %v", err)
	}

	if out.Value != 7 {
		t.Errorf("value = %d, want 7", out.Value)
	}
}
