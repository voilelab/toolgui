package tgcomp

import (
	"github.com/voilelab/toolgui/toolgui/tgcomp/tcchat"
	"github.com/voilelab/toolgui/toolgui/tgframe"
)

// ChatMessage draws a message of a chat, sent by role, and returns the
// container its content goes in.
func ChatMessage(c *tgframe.Container, role string,
	conf ...*ChatMessageConf) *tgframe.Container {
	return tcchat.ChatMessage(c, role, conf...)
}

// ChatMessageConf is the configuration for the ChatMessage component.
type ChatMessageConf = tcchat.ChatMessageConf
