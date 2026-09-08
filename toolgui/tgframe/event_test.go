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

// What the frontend puts on the wire reads back as the right fields, and
// marshalling keeps an empty values there while leaving an event that never
// had one without it. A multi-valued payload carries no value of its own, so
// what comes back out is not what went in: Go's zero fills the field the
// frontend omitted, which is harmless only because values is what ApplyState
// reads once it is present.
func TestEventSelectMarshalShape(t *testing.T) {
	for _, tt := range []struct {
		name string
		sent string
		want string
	}{
		{"single", `{"id":"a","value":2}`, `{"id":"a","value":2}`},
		{"multi", `{"id":"a","values":[0,2]}`, `{"id":"a","value":0,"values":[0,2]}`},
		{"empty multi", `{"id":"a","values":[]}`, `{"id":"a","value":0,"values":[]}`},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var event EventSelect
			if err := json.Unmarshal([]byte(tt.sent), &event); err != nil {
				t.Fatalf("Unmarshal: %v", err)
			}

			bs, err := json.Marshal(&event)
			if err != nil {
				t.Fatalf("Marshal: %v", err)
			}

			if string(bs) != tt.want {
				t.Errorf("Marshal = %s, want %s", bs, tt.want)
			}
		})
	}
}
