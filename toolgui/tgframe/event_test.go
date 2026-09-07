package tgframe

import (
	"testing"
)

func TestParseEventIframe(t *testing.T) {
	event, err := ParseEvent([]byte(`{"type":"iframe","id":"my_iframe","value":{"clicked":true}}`))
	if err != nil {
		t.Fatalf("ParseEvent: %v", err)
	}

	iframeEvent, ok := event.(*EventIframe)
	if !ok {
		t.Fatalf("got %T, want *EventIframe", event)
	}

	if iframeEvent.ID != "my_iframe" {
		t.Errorf("ID = %q, want %q", iframeEvent.ID, "my_iframe")
	}

	// The value lands in the state under the iframe's own id.
	state := NewState()
	iframeEvent.ApplyState(state)

	var out struct {
		Clicked bool `json:"clicked"`
	}
	if err := state.GetObject("my_iframe", &out); err != nil {
		t.Fatalf("GetObject: %v", err)
	}

	if !out.Clicked {
		t.Error("clicked = false, want true")
	}
}

// An iframe event must not be able to stand in for a button click.
func TestParseEventIframeDoesNotSetClickID(t *testing.T) {
	event, err := ParseEvent([]byte(`{"type":"iframe","id":"some_button","value":null}`))
	if err != nil {
		t.Fatalf("ParseEvent: %v", err)
	}

	state := NewState()
	event.ApplyState(state)

	if state.GetClickID() != "" {
		t.Errorf("ClickID = %q, want empty", state.GetClickID())
	}
}

func TestParseEventCustom(t *testing.T) {
	event, err := ParseEvent([]byte(`{"type":"custom","id":"my_plugin","value":{"clicked":true}}`))
	if err != nil {
		t.Fatalf("ParseEvent: %v", err)
	}

	customEvent, ok := event.(*EventCustom)
	if !ok {
		t.Fatalf("got %T, want *EventCustom", event)
	}

	if customEvent.ID != "my_plugin" {
		t.Errorf("ID = %q, want %q", customEvent.ID, "my_plugin")
	}

	state := NewState()
	customEvent.ApplyState(state)

	var out struct {
		Clicked bool `json:"clicked"`
	}
	if err := state.GetObject("my_plugin", &out); err != nil {
		t.Fatalf("GetObject: %v", err)
	}

	if !out.Clicked {
		t.Error("clicked = false, want true")
	}
}

// A frontend built before the event was renamed still sends "iframe", and an
// app it talks to has to keep understanding it.
func TestParseEventIframeIsTheCustomEvent(t *testing.T) {
	event, err := ParseEvent([]byte(`{"type":"iframe","id":"my_iframe","value":1}`))
	if err != nil {
		t.Fatalf("ParseEvent: %v", err)
	}

	if _, ok := event.(*EventCustom); !ok {
		t.Fatalf("got %T, want *EventCustom", event)
	}
}
