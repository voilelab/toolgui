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
