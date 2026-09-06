package tcmisc

import (
	"encoding/json"
	"testing"

	"github.com/voilelab/toolgui/toolgui/tgframe"
)

// addIframe runs the given call against a container and returns the json the
// container would have sent to the client for the added component.
func addIframe(t *testing.T, add func(c *tgframe.Container)) map[string]any {
	t.Helper()

	var packs []tgframe.NotifyPack
	container := tgframe.NewContainer("test", func(pack tgframe.NotifyPack) {
		packs = append(packs, pack)
	})

	add(container)

	if len(packs) != 1 {
		t.Fatalf("got %d packs, want 1", len(packs))
	}

	bs, err := json.Marshal(packs[0])
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var out struct {
		Component map[string]any `json:"component"`
	}
	if err := json.Unmarshal(bs, &out); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	return out.Component
}

func TestIframeDefaultSize(t *testing.T) {
	props := addIframe(t, func(c *tgframe.Container) {
		Iframe(c, "<b>hi</b>", true)
	})

	if props["width"] != defaultIframeWidth {
		t.Errorf("width = %v, want %v", props["width"], defaultIframeWidth)
	}

	if props["height"] != defaultIframeHeight {
		t.Errorf("height = %v, want %v", props["height"], defaultIframeHeight)
	}
}

func TestIframeWithConf(t *testing.T) {
	props := addIframe(t, func(c *tgframe.Container) {
		IframeWithConf(c, "<b>hi</b>", &IframeConf{
			Script: true,
			Width:  "300px",
			Height: "400px",
			ID:     "my_iframe",
		})
	})

	if props["width"] != "300px" {
		t.Errorf("width = %v, want 300px", props["width"])
	}

	if props["height"] != "400px" {
		t.Errorf("height = %v, want 400px", props["height"])
	}

	if props["id"] != "iframe_component_my_iframe" {
		t.Errorf("id = %v, want iframe_component_my_iframe", props["id"])
	}

	if props["script"] != true {
		t.Errorf("script = %v, want true", props["script"])
	}
}

// The state key IframeValue reads must be the component id the frontend puts
// on the event, not the bare user id.
func TestIframeValueRoundTrip(t *testing.T) {
	props := addIframe(t, func(c *tgframe.Container) {
		IframeWithID(c, "<b>hi</b>", true, "my_iframe")
	})

	// What the frontend sends back: the iframe's own component id.
	event := &tgframe.EventIframe{
		ID:    props["id"].(string),
		Value: map[string]any{"clicked": true},
	}

	state := tgframe.NewState()
	event.ApplyState(state)

	var out struct {
		Clicked bool `json:"clicked"`
	}
	if err := IframeValue(state, "my_iframe", &out); err != nil {
		t.Fatalf("IframeValue: %v", err)
	}

	if !out.Clicked {
		t.Error("clicked = false, want true")
	}
}
