package tgframe

import (
	"encoding/json/jsontext"
	"errors"
	"fmt"
	"log"

	"github.com/voilelab/toolgui/toolgui/tgjson"
	"github.com/voilelab/toolgui/toolgui/tgutil"
)

// Event is the state change made by app user
type Event interface {
	ApplyState(state *State)
}

// ErrUnknownComponentID is what an event writing under an id no component on
// the page owns is turned away with.
var ErrUnknownComponentID = errors.New("unknown component id")

// stateWriter is an event that writes the state under an id the client fills
// in. [Session] checks the id before applying the event, so the ids a client
// may write to are the page's own rather than whatever it cares to name.
type stateWriter interface {
	// validateState returns an error when the event writes under an id the
	// page is not showing.
	validateState(state *State) error
}

// validateEvent reports whether event may be applied to state. An event that
// writes nothing the client names -- a click, a rerun -- is always fine.
func validateEvent(event Event, state *State) error {
	writer, ok := event.(stateWriter)
	if !ok {
		return nil
	}

	return writer.validateState(state)
}

// validateComponentID returns an error unless id names a component the last
// finished run drew.
//
// Before any run finishes the set is empty, so every id is turned away. That
// is the right answer rather than a gap: a client has been shown nothing to
// write to yet, and the first event a frontend sends on connect is the empty
// rerun event, which names no id. Letting unknown ids through until the first
// run finishes would also leave the hole this check closes wide open, since
// an event arriving before the run it started has finished cuts that run, and
// a cut run records no ids.
func validateComponentID(state *State, id string) error {
	if state.HasComponentID(id) {
		return nil
	}

	return tgutil.Errorf("%w: `%s`", ErrUnknownComponentID, id)
}

// EventType is the type of event
type EventType string

const (
	EventEmptyName  EventType = ""
	EventClickName  EventType = "click"
	EventInputName  EventType = "input"
	EventSelectName EventType = "select"
	EventFormName   EventType = "form"

	// EventCustomName is what a component that renders itself sends its value
	// back with. The frontend fills the id in with the component's own id, and
	// the server only takes an id the page drew, so such a component writes
	// its own state and nothing else.
	EventCustomName EventType = "custom"

	// EventIframeName is the name the custom event was introduced under.
	//
	// Deprecated: use [EventCustomName].
	EventIframeName EventType = "iframe"
)

type EventStruct struct {
	Type EventType `json:"type"`
}

// maxFormDepth is how deep a form event may nest. Every level has its subtree
// unmarshalled again, so a message costs its size times its depth to read: a
// form of forms is what buys an attacker more work than it took to send. A
// form holds inputs rather than forms, so one level is the real case.
const maxFormDepth = 32

// ErrFormTooDeep is the error that a form event nests past [maxFormDepth].
var ErrFormTooDeep = errors.New("form event nested too deep")

// ParseEvent parses an event off the wire.
func ParseEvent(data []byte) (Event, error) {
	return parseEvent(data, 0)
}

// parseEvent parses one event nested depth forms deep.
func parseEvent(data []byte, depth int) (Event, error) {
	var event EventStruct
	err := tgjson.Unmarshal(data, &event)
	if err != nil {
		return nil, err
	}

	switch event.Type {
	case EventEmptyName:
		return &EventEmpty{}, nil
	case EventClickName:
		var eventClick EventClick
		err = tgjson.Unmarshal(data, &eventClick)
		if err != nil {
			return nil, err
		}
		return &eventClick, nil
	case EventInputName:
		var eventInput EventInput
		err = tgjson.Unmarshal(data, &eventInput)
		if err != nil {
			return nil, err
		}
		return &eventInput, nil
	case EventSelectName:
		var eventSelect EventSelect
		err = tgjson.Unmarshal(data, &eventSelect)
		if err != nil {
			return nil, err
		}
		return &eventSelect, nil
	case EventFormName:
		if depth >= maxFormDepth {
			return nil, fmt.Errorf("%w: at most %d levels", ErrFormTooDeep,
				maxFormDepth)
		}

		var eventForm struct {
			Events []jsontext.Value `json:"events"`
		}
		err = tgjson.Unmarshal(data, &eventForm)
		if err != nil {
			return nil, err
		}
		events := []Event{}
		for _, event := range eventForm.Events {
			parsedEvent, err := parseEvent(event, depth+1)
			if err != nil {
				// A child the parser doesn't understand is dropped, but one
				// nested too deep condemns the message: what it costs to read
				// is the tree underneath it, and nothing has to read that.
				if errors.Is(err, ErrFormTooDeep) {
					return nil, err
				}

				log.Printf("failed to parse event: %v", err)
				continue
			}
			events = append(events, parsedEvent)
		}
		return &EventForm{Events: events}, nil
	case EventCustomName, EventIframeName:
		var eventCustom EventCustom
		err = tgjson.Unmarshal(data, &eventCustom)
		if err != nil {
			return nil, err
		}
		return &eventCustom, nil
	default:
		return nil, fmt.Errorf("unknown event type: %s", event.Type)
	}
}

// EventEmpty is the event of an empty event, rerun button will send this
type EventEmpty struct {
}

func (e *EventEmpty) ApplyState(*State) {
}

// EventClick is the event of a button click event
type EventClick struct {
	ID string `json:"id"`
}

func (e *EventClick) ApplyState(state *State) {
	state.SetClickID(e.ID)
}

// EventInput is the event of a input event
// it's used for all input components
type EventInput struct {
	ID    string `json:"id"`
	Value any    `json:"value"`
}

func (e *EventInput) ApplyState(state *State) {
	state.Set(e.ID, e.Value)
}

func (e *EventInput) validateState(state *State) error {
	return validateComponentID(state, e.ID)
}

// EventSelect is the event of a select event
// it's used for select/radio/multiselect component
type EventSelect struct {
	ID    string `json:"id"`
	Value int    `json:"value"`

	// Values carries the selection of a component that takes more than one,
	// and is absent from a single-valued one's payload. It is a field of its
	// own rather than a Value that may be either shape so that a frontend
	// built before multi-selection existed keeps working unchanged.
	Values []int `json:"values,omitzero"`
}

func (e *EventSelect) ApplyState(state *State) {
	// An empty selection still arrives as [], so the nil check is what tells
	// a multi-valued payload from a single-valued one, not the length.
	if e.Values != nil {
		state.Set(e.ID, e.Values)
		return
	}

	state.Set(e.ID, e.Value)
}

func (e *EventSelect) validateState(state *State) error {
	return validateComponentID(state, e.ID)
}

// EventForm is the event of a form event
// it's used for form component
type EventForm struct {
	Events []Event `json:"events"`
}

func (e *EventForm) ApplyState(state *State) {
	for _, event := range e.Events {
		event.ApplyState(state)
	}
}

// validateState turns the whole form away when one of its events names an id
// the page is not showing. A form is applied in one go, so it is checked in
// one go too: half a form written is a state no submit ever produced.
func (e *EventForm) validateState(state *State) error {
	for _, event := range e.Events {
		if err := validateEvent(event, state); err != nil {
			return tgutil.Errorf("%w", err)
		}
	}

	return nil
}

// EventCustom carries an arbitrary value from a component that renders
// itself, such as an iframe or a plugin. The ID is filled in by the frontend
// with the component's own id, and the server takes only the ids the page
// drew, so such a component writes its own state key and cannot make up one
// of its own.
type EventCustom struct {
	ID    string `json:"id"`
	Value any    `json:"value"`
}

func (e *EventCustom) ApplyState(state *State) {
	state.Set(e.ID, e.Value)
}

func (e *EventCustom) validateState(state *State) error {
	return validateComponentID(state, e.ID)
}

// EventIframe is the name [EventCustom] was introduced under.
//
// Deprecated: use [EventCustom].
type EventIframe = EventCustom
