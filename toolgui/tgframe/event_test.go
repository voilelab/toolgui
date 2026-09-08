package tgframe

import (
	"encoding/json"
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

// The single-valued payload a select or a radio sends is what it always was,
// and a frontend that has never heard of a multi-valued one keeps working.
func TestParseEventSelectSingleValue(t *testing.T) {
	event, err := ParseEvent([]byte(`{"type":"select","id":"select_component_Fruit","value":2}`))
	if err != nil {
		t.Fatalf("ParseEvent: %v", err)
	}

	selectEvent, ok := event.(*EventSelect)
	if !ok {
		t.Fatalf("got %T, want *EventSelect", event)
	}

	if selectEvent.Values != nil {
		t.Errorf("Values = %v, want nil", selectEvent.Values)
	}

	state := NewState()
	selectEvent.ApplyState(state)

	idx := state.GetInt("select_component_Fruit")
	if idx == nil || *idx != 2 {
		t.Fatalf("state = %v, want 2", idx)
	}
}

func TestParseEventSelectMultiValue(t *testing.T) {
	event, err := ParseEvent([]byte(
		`{"type":"select","id":"multiselect_component_Fruit","values":[0,2]}`))
	if err != nil {
		t.Fatalf("ParseEvent: %v", err)
	}

	selectEvent, ok := event.(*EventSelect)
	if !ok {
		t.Fatalf("got %T, want *EventSelect", event)
	}

	state := NewState()
	selectEvent.ApplyState(state)

	var got []int
	if err := state.GetObject("multiselect_component_Fruit", &got); err != nil {
		t.Fatalf("GetObject: %v", err)
	}

	if len(got) != 2 || got[0] != 0 || got[1] != 2 {
		t.Fatalf("state = %v, want [0 2]", got)
	}
}

// Deselecting the last item sends an empty list, which has to stay a
// selection of nothing rather than fall through to the single value's zero.
func TestParseEventSelectEmptyMultiValue(t *testing.T) {
	event, err := ParseEvent([]byte(
		`{"type":"select","id":"multiselect_component_Fruit","values":[]}`))
	if err != nil {
		t.Fatalf("ParseEvent: %v", err)
	}

	state := NewState()
	event.ApplyState(state)

	var got []int
	if err := state.GetObject("multiselect_component_Fruit", &got); err != nil {
		t.Fatalf("GetObject: %v", err)
	}

	if got == nil || len(got) != 0 {
		t.Fatalf("state = %v, want an empty selection", got)
	}
}

// A multi-valued event marshals back to the shape the frontend sends, and a
// single-valued one does not grow a values field it never had.
func TestEventSelectRoundTrip(t *testing.T) {
	for _, tt := range []struct {
		name string
		data string
	}{
		{"single", `{"id":"a","value":2}`},
		{"multi", `{"id":"a","value":0,"values":[0,2]}`},
		{"empty multi", `{"id":"a","value":0,"values":[]}`},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var event EventSelect
			if err := json.Unmarshal([]byte(tt.data), &event); err != nil {
				t.Fatalf("Unmarshal: %v", err)
			}

			bs, err := json.Marshal(&event)
			if err != nil {
				t.Fatalf("Marshal: %v", err)
			}

			if string(bs) != tt.data {
				t.Errorf("round trip = %s, want %s", bs, tt.data)
			}
		})
	}
}
