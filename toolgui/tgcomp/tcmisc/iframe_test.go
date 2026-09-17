package tcmisc

import (
	"testing"

	"github.com/voilelab/toolgui/toolgui/tgframe"
	"github.com/voilelab/toolgui/toolgui/tgjson"
)

// addIframe runs the given call against a container and returns the json the
// container would have sent to the client for the added component.
func addIframe(t *testing.T, add func(c *tgframe.Container)) map[string]any {
	t.Helper()

	var packs []tgframe.NotifyPack
	container := tgframe.NewContainer("test", tgframe.NewState(), func(pack tgframe.NotifyPack) {
		packs = append(packs, pack)
	})

	add(container)

	if len(packs) != 1 {
		t.Fatalf("got %d packs, want 1", len(packs))
	}

	bs, err := tgjson.Marshal(packs[0])
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var out struct {
		Component map[string]any `json:"component"`
	}
	if err := tgjson.Unmarshal(bs, &out); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	return out.Component
}

// drawWithState runs the given call against a container backed by state, the
// way a rerun draws a page whose widgets already have values, and returns the
// names of the components it drew.
func drawWithState(t *testing.T, state *tgframe.State, add func(c *tgframe.Container)) []string {
	t.Helper()

	var names []string
	container := tgframe.NewContainer("test", state, func(pack tgframe.NotifyPack) {
		bs, err := tgjson.Marshal(pack)
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}

		var out struct {
			Component struct {
				Name string `json:"name"`
			} `json:"component"`
		}
		if err := tgjson.Unmarshal(bs, &out); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}

		names = append(names, out.Component.Name)
	})

	add(container)

	return names
}

// stateWithValue replays what the frontend sends for the component drawn by
// add: it draws once to learn the id, then puts value under it.
func stateWithValue(t *testing.T, value any, add func(c *tgframe.Container)) *tgframe.State {
	t.Helper()

	props := addIframe(t, add)

	state := tgframe.NewState()
	event := &tgframe.EventCustom{ID: props["id"].(string), Value: value}
	event.ApplyState(state)

	return state
}

func TestIframeDefaultSize(t *testing.T) {
	props := addIframe(t, func(c *tgframe.Container) {
		Iframe(c, "<b>hi</b>", &IframeConf{Script: true})
	})

	if props["width"] != defaultIframeWidth {
		t.Errorf("width = %v, want %v", props["width"], defaultIframeWidth)
	}

	if props["height"] != defaultIframeHeight {
		t.Errorf("height = %v, want %v", props["height"], defaultIframeHeight)
	}
}

func TestIframeConfDrivesTheProps(t *testing.T) {
	props := addIframe(t, func(c *tgframe.Container) {
		Iframe(c, "<b>hi</b>", &IframeConf{
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

// A value the guest sent comes back typed, keyed by the iframe's own
// component id rather than by the bare user id.
func TestIframeValueRoundTrip(t *testing.T) {
	const html = "<b>hi</b>"
	conf := &IframeConf{Script: true, ID: "my_iframe"}

	type picked struct {
		Clicked bool `json:"clicked"`
	}

	state := stateWithValue(t, map[string]any{"clicked": true}, func(c *tgframe.Container) {
		Iframe(c, html, conf)
	})

	var got *picked
	drawWithState(t, state, func(c *tgframe.Container) {
		got = IframeValue[picked](c, html, conf)
	})

	if got == nil {
		t.Fatal("value = nil, want the value the guest sent")
	}

	if !got.Clicked {
		t.Error("clicked = false, want true")
	}
}

// Before the guest sends anything there is no value, which is not the same as
// a value that is the zero T.
func TestIframeValueBeforeTheFirstUpdate(t *testing.T) {
	type picked struct {
		Clicked bool `json:"clicked"`
	}

	var got *picked
	drawWithState(t, tgframe.NewState(), func(c *tgframe.Container) {
		got = IframeValue[picked](c, "<b>hi</b>", &IframeConf{Script: true})
	})

	if got != nil {
		t.Errorf("value = %v, want nil before the guest sends one", got)
	}
}

// Reading is not drawing: a page that only wants the value does not put a
// second iframe on the screen.
func TestIframeValueDrawsNothing(t *testing.T) {
	names := drawWithState(t, tgframe.NewState(), func(c *tgframe.Container) {
		_ = IframeValue[struct{}](c, "<b>hi</b>", &IframeConf{Script: true})
	})

	if len(names) != 0 {
		t.Errorf("drew %v, want nothing", names)
	}
}
