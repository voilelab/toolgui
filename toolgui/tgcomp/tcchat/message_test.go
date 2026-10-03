package tcchat

import (
	"testing"

	"github.com/voilelab/toolgui/toolgui/tgframe"
	"github.com/voilelab/toolgui/toolgui/tgjson"
)

// chatPacks runs add against a container and returns the props of every pack
// it sent, in order.
func chatPacks(t *testing.T, add func(c *tgframe.Container)) []map[string]any {
	t.Helper()
	return chatPacksIn(t, "test", add)
}

// chatPacksIn is chatPacks in a root container with the given id.
func chatPacksIn(t *testing.T, id string, add func(c *tgframe.Container)) []map[string]any {
	t.Helper()

	var props []map[string]any
	c := tgframe.NewContainer(id, tgframe.NewState(), func(pack tgframe.NotifyPack) {
		bs, err := tgjson.Marshal(pack)
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}

		var out struct {
			Component map[string]any `json:"component"`
		}
		if err := tgjson.Unmarshal(bs, &out); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}

		props = append(props, out.Component)
	})

	add(c)
	return props
}

func TestChatMessageAvatar(t *testing.T) {
	cases := []struct {
		role, avatar   string
		kind, src, txt string
	}{
		{role: "user", kind: kindUser},
		{role: "Human", kind: kindUser},
		{role: "assistant", kind: kindAssistant},
		{role: "AI", kind: kindAssistant},
		{role: "tool", kind: kindOther, txt: "T"},
		{role: "élan", kind: kindOther, txt: "É"},
		{role: "", kind: kindOther},
		{role: "user", avatar: "🦖", kind: kindUser, txt: "🦖"},
		{role: "tool", avatar: "https://x/a.png", kind: kindOther, src: "https://x/a.png"},
		{role: "ai", avatar: "/static/bot.png", kind: kindAssistant, src: "/static/bot.png"},
		{role: "ai", avatar: "data:image/png;base64,AA", kind: kindAssistant, src: "data:image/png;base64,AA"},
	}

	for _, tc := range cases {
		comp := newChatMessageComponent(tc.role, tc.avatar)
		if comp.Kind != tc.kind || comp.AvatarSrc != tc.src || comp.AvatarText != tc.txt {
			t.Errorf("(%q, %q) = kind %q src %q text %q, want %q %q %q",
				tc.role, tc.avatar, comp.Kind, comp.AvatarSrc, comp.AvatarText,
				tc.kind, tc.src, tc.txt)
		}
	}
}

func TestChatMessageContainer(t *testing.T) {
	var inner *tgframe.Container
	props := chatPacks(t, func(c *tgframe.Container) {
		inner = ChatMessage(c, "assistant", &ChatMessageConf{Base: tgframe.Base{ID: "reply"}})
		inner.AddComponent(&tgframe.BaseComponent{Name: "text_component"})
	})

	if len(props) != 3 {
		t.Fatalf("got %d packs, want message, its container, the text", len(props))
	}

	if props[0]["name"] != chatMessageComponentName || props[0]["role"] != "assistant" ||
		props[0]["id"] != "chat_message_component_reply" {
		t.Errorf("message = %v", props[0])
	}

	want := "chat_message_component_reply_inner"
	if props[1]["id"] != want || inner.GetID() != want {
		t.Errorf("container id = %v, want %s", props[1]["id"], want)
	}
}
