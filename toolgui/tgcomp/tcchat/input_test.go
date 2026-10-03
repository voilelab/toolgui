package tcchat

import (
	"testing"

	"github.com/voilelab/toolgui/toolgui/tgframe"
)

const inputID = "chat_input_component_Say something"

// send applies the event the client sends for msg.
func send(state *tgframe.State, msg string) {
	(&tgframe.EventForm{Events: []tgframe.Event{
		&tgframe.EventInput{ID: inputID, Value: msg},
		&tgframe.EventClick{ID: inputID},
	}}).ApplyState(state)
}

func TestChatInputSend(t *testing.T) {
	state := tgframe.NewState()
	c := tgframe.NewContainer(tgframe.MainContainerID, state, func(tgframe.NotifyPack) {})

	if msg, ok := ChatInput(c, "Say something"); ok || msg != "" {
		t.Errorf("before a send = (%q, %v), want nothing", msg, ok)
	}

	send(state, "hello")
	c = tgframe.NewContainer(tgframe.MainContainerID, state, func(tgframe.NotifyPack) {})
	if msg, ok := ChatInput(c, "Say something"); !ok || msg != "hello" {
		t.Errorf("on the send = (%q, %v), want hello", msg, ok)
	}

	// Another event's run is not the send's.
	state.SetClickID("button_component_other")
	c = tgframe.NewContainer(tgframe.MainContainerID, state, func(tgframe.NotifyPack) {})
	if msg, ok := ChatInput(c, "Say something"); ok || msg != "" {
		t.Errorf("after the send = (%q, %v), want nothing", msg, ok)
	}
}

// A click on the input without a message, or a message from another
// component's click, is no send.
func TestChatInputNeedsBoth(t *testing.T) {
	state := tgframe.NewState()
	state.SetClickID(inputID)
	c := tgframe.NewContainer(tgframe.MainContainerID, state, func(tgframe.NotifyPack) {})
	if _, ok := ChatInput(c, "Say something"); ok {
		t.Error("click alone counted as a send")
	}

	state.Set(inputID, "stale")
	state.SetClickID("button_component_other")
	if _, ok := ChatInput(c, "Say something"); ok {
		t.Error("value alone counted as a send")
	}

	if _, ok := state.Get[string](inputID); ok {
		t.Error("message outlived the run")
	}
}

func TestChatInputPinned(t *testing.T) {
	props := chatPacks(t, func(c *tgframe.Container) {
		ChatInput(c, "Say something")
	})
	if props[0]["pinned"] != false {
		t.Errorf("pinned = %v in a test container, want false", props[0]["pinned"])
	}

	props = chatPacksIn(t, tgframe.MainContainerID, func(c *tgframe.Container) {
		ChatInput(c, "Say something")
		ChatInput(c.AddContainer("box"), "Nested", &ChatInputConf{MaxLength: 10})
	})

	pinned := []any{}
	for _, p := range props {
		if p["name"] == chatInputComponentName {
			pinned = append(pinned, p["pinned"])
		}
	}

	if len(pinned) != 2 || pinned[0] != true || pinned[1] != false {
		t.Errorf("pinned = %v, want main pinned, nested not", pinned)
	}
}
