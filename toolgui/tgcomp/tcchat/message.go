// Package tcchat provides the components a chat is made of.
package tcchat

import (
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/voilelab/toolgui/toolgui/tgframe"
)

var _ tgframe.Component = &chatMessageComponent{}

const chatMessageComponentName = "chat_message_component"

// The kinds a role is drawn as. The user and the assistant get an icon of
// their own; any other role is drawn by its first letter.
const (
	kindUser      = "user"
	kindAssistant = "assistant"
	kindOther     = "other"
)

type chatMessageComponent struct {
	*tgframe.BaseComponent

	Role string `json:"role"`
	Kind string `json:"kind"`

	// One of the two is set, or neither: an image, or a text such as an
	// emoji. Neither leaves the kind's icon.
	AvatarSrc  string `json:"avatar_src"`
	AvatarText string `json:"avatar_text"`
}

func newChatMessageComponent(role, avatar string) *chatMessageComponent {
	comp := &chatMessageComponent{
		BaseComponent: &tgframe.BaseComponent{
			Name: chatMessageComponentName,
		},
		Role: role,
		Kind: roleKind(role),
	}

	switch {
	case isImageURL(avatar):
		comp.AvatarSrc = avatar
	case avatar != "":
		comp.AvatarText = avatar
	case comp.Kind == kindOther:
		comp.AvatarText = initial(role)
	}

	return comp
}

func roleKind(role string) string {
	switch strings.ToLower(role) {
	case "user", "human":
		return kindUser
	case "assistant", "ai":
		return kindAssistant
	default:
		return kindOther
	}
}

func isImageURL(s string) bool {
	for _, prefix := range []string{"http://", "https://", "data:image/", "/", "./"} {
		if strings.HasPrefix(s, prefix) {
			return true
		}
	}

	return false
}

// initial is the first letter of role, upper-cased.
func initial(role string) string {
	r, _ := utf8.DecodeRuneInString(strings.TrimSpace(role))
	if r == utf8.RuneError {
		return ""
	}

	return string(unicode.ToUpper(r))
}

// ChatMessageConf is the configuration for the ChatMessage component.
type ChatMessageConf struct {
	tgframe.Base

	// Avatar is an emoji, or an image URL: http, https, data:image or a path
	// starting with / or ./. Default is an icon for the user and the
	// assistant, and the role's first letter for any other role.
	Avatar string
}

// ChatMessage draws a message of a chat, sent by role, and returns the
// container its content goes in. "user" and "human" are drawn as the user,
// "assistant" and "ai" as the assistant; any other role is drawn by name.
func ChatMessage(c *tgframe.Container, role string,
	conf ...*ChatMessageConf) *tgframe.Container {
	cf := tgframe.OneConf("ChatMessage", conf)

	comp := newChatMessageComponent(role, cf.Avatar)
	tgframe.SetConfIDIn(c, comp, cf)

	c.AddComponent(comp)
	return c.AddContainerTo(comp, "inner", 0)
}
