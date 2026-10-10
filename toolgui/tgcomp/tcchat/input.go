package tcchat

import (
	"github.com/voilelab/toolgui/toolgui/tgcomp/tcutil"
	"github.com/voilelab/toolgui/toolgui/tgframe"
)

var _ tgframe.Component = &chatInputComponent{}

const chatInputComponentName = "chat_input_component"

// mainContainerID is the id of the page's main container, the one a chat
// input is pinned to the bottom of.
var mainContainerID = tgframe.ContainerComponentName + "_" + tgframe.MainContainerID

type chatInputComponent struct {
	*tgframe.BaseComponent

	Placeholder string `json:"placeholder"`
	MaxLength   int    `json:"max_length"`
	Disabled    bool   `json:"disabled"`

	// Pinned draws it at the bottom of the page instead of where it is
	// written.
	Pinned bool `json:"pinned"`
}

func newChatInputComponent(placeholder string) *chatInputComponent {
	return &chatInputComponent{
		BaseComponent: &tgframe.BaseComponent{
			Name: chatInputComponentName,
			ID:   tcutil.NormalID(chatInputComponentName, placeholder),
		},
		Placeholder: placeholder,
	}
}

// ChatInputConf is the configuration for the ChatInput component.
//
//tgcomp:export
type ChatInputConf struct {
	tgframe.Base

	// MaxLength is the most characters a message may have. 0 is no limit.
	MaxLength int

	// Disabled keeps the app user from sending, e.g. while a reply streams.
	Disabled bool
}

// ChatInput draws a box to send a chat message, and returns the message and
// true on the run its send starts. Every other run returns "" and false. The
// box is emptied once a message is sent.
//
// Written in the page's main container, it is pinned to the bottom of the
// page wherever it is written, so it may come before the messages it adds.
// In any other container it stays where it is written.
//
//tgcomp:export
func ChatInput(c *tgframe.Container, placeholder string,
	conf ...*ChatInputConf) (string, bool) {
	cf := tgframe.OneConf("ChatInput", conf)

	comp := newChatInputComponent(placeholder)
	comp.MaxLength = cf.MaxLength
	comp.Disabled = cf.Disabled
	comp.Pinned = c.GetID() == mainContainerID
	tgframe.SetConfIDIn(c, comp, cf)

	c.AddComponent(comp)

	// The message is the send's alone, so it does not outlive the run.
	msg, ok := c.State.Get[string](comp.ID)
	c.State.Delete(comp.ID)

	if !ok || c.State.GetClickID() != comp.ID {
		return "", false
	}

	return msg, true
}
