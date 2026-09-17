package tgframe

import (
	"encoding/json/jsontext"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/voilelab/toolgui/toolgui/tgjson"
)

// sameJSON reports whether two encodings carry the same value. The tests
// compare what a pack means, not the bytes a codec happened to choose, so a
// change of member order or number formatting is not a failure.
func sameJSON(t *testing.T, got, want []byte) bool {
	t.Helper()

	canon := func(what string, bs []byte) string {
		v := jsontext.Value(bs)
		if err := v.Canonicalize(); err != nil {
			t.Fatalf("canonicalize %s: %v", what, err)
		}

		return string(v)
	}

	return canon("got", got) == canon("want", want)
}

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

	idx, ok := state.GetNumber[int]("select_component_Fruit")
	if !ok || idx != 2 {
		t.Fatalf("state = %v, %v, want 2, true", idx, ok)
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
			if err := tgjson.Unmarshal([]byte(tt.sent), &event); err != nil {
				t.Fatalf("Unmarshal: %v", err)
			}

			bs, err := tgjson.Marshal(&event)
			if err != nil {
				t.Fatalf("Marshal: %v", err)
			}

			if !sameJSON(t, bs, []byte(tt.want)) {
				t.Errorf("Marshal = %s, want %s", bs, tt.want)
			}
		})
	}
}

// The codec's strictness reaches the wire path an event actually arrives on:
// a duplicate key or invalid UTF-8 is rejected rather than read as whichever
// of the two values the decoder happened to keep.
func TestParseEventRejectsMalformed(t *testing.T) {
	for _, tc := range []struct {
		name string
		data string
	}{
		{"duplicate id", `{"type":"click","id":"a","id":"b"}`},
		{"duplicate type", `{"type":"click","type":"input"}`},
		{"invalid utf-8", "{\"type\":\"click\",\"id\":\"a\xffb\"}"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			event, err := ParseEvent([]byte(tc.data))
			if err == nil {
				t.Errorf("ParseEvent(%q) = %#v, want an error", tc.data, event)
			}
		})
	}
}

// nestedFormEvent builds depth forms wrapped around one input, with pad bytes
// of padding in the input's value.
func nestedFormEvent(depth, pad int) []byte {
	event := `{"type":"input","id":"a","value":"` + strings.Repeat("x", pad) + `"}`
	for range depth {
		event = `{"type":"form","events":[` + event + `]}`
	}

	return []byte(event)
}

// Every level of a form has its subtree unmarshalled again, so a message costs
// its size times its depth to read. The depth is what a cap has to be put on:
// without one a single message near the size limit buys orders of magnitude
// more work than it took to send, and the process is read out of memory.
func TestParseEventRejectsDeepForm(t *testing.T) {
	for _, tc := range []struct {
		name  string
		depth int
		want  bool
	}{
		{"at the cap", maxFormDepth, false},
		{"one past it", maxFormDepth + 1, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ParseEvent(nestedFormEvent(tc.depth, 0))

			if got := errors.Is(err, ErrFormTooDeep); got != tc.want {
				t.Errorf("ParseEvent: err = %v, ErrFormTooDeep = %v, want %v",
					err, got, tc.want)
			}
		})
	}
}

// What the cap buys, with the size cap the web executor puts on a message: a
// message as big as one may be, nested as deep as one may be, costs a bounded
// amount to read. It used to cost its size times its depth, which is how a
// message well under the size limit read the process out of memory.
func TestParseEventDeepFormCostIsBounded(t *testing.T) {
	// The size cap lives in tgexec, which this package cannot import.
	const maxMessageSize = 1 << 20

	start := time.Now()
	_, err := ParseEvent(nestedFormEvent(maxFormDepth, maxMessageSize))
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("ParseEvent: %v", err)
	}

	// Two orders of magnitude over what it measures, so a slow machine does
	// not fail the test and a return to the old cost does.
	if elapsed > 10*time.Second {
		t.Errorf("took %v, want a bounded cost", elapsed)
	}
}

// A form is a form of inputs, and the cap is nowhere near what a page draws.
func TestParseEventForm(t *testing.T) {
	event, err := ParseEvent([]byte(
		`{"type":"form","events":[` +
			`{"type":"input","id":"a","value":"1"},` +
			`{"type":"select","id":"b","value":2}]}`))
	if err != nil {
		t.Fatalf("ParseEvent: %v", err)
	}

	formEvent, ok := event.(*EventForm)
	if !ok {
		t.Fatalf("got %T, want *EventForm", event)
	}

	if len(formEvent.Events) != 2 {
		t.Fatalf("got %d events, want 2", len(formEvent.Events))
	}

	state := NewState()
	formEvent.ApplyState(state)

	if got := state.values["a"]; got != "1" {
		t.Errorf("state[a] = %v, want %q", got, "1")
	}

	if got := state.values["b"]; got != 2 {
		t.Errorf("state[b] = %v, want 2", got)
	}
}
