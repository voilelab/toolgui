package tcmisc

import (
	"github.com/voilelab/toolgui/toolgui/tgcomp/tcutil"
	"github.com/voilelab/toolgui/toolgui/tgframe"
)

var _ tgframe.Component = &messageComponent{}
var messageComponentName = "message_component"

type messageComponent struct {
	*tgframe.BaseComponent
	Title string       `json:"title"`
	Body  string       `json:"body"`
	Color tcutil.Color `json:"color"`
}

func newMessageComponent(body string) *messageComponent {
	return &messageComponent{
		BaseComponent: &tgframe.BaseComponent{
			Name: messageComponentName,
		},
		Body: body,
	}
}

// MessageConf provide extra config for Message Component.
type MessageConf struct {
	tgframe.Base

	// Title is the title of the message. Optional.
	Title string

	// Color is the color of the message. Default is tcutil.ColorNull.
	// The colored variants set it themselves.
	Color tcutil.Color
}

// Message is a component that displays a message.
func Message(c *tgframe.Container, text string, conf ...*MessageConf) {
	message(c, text, tgframe.OneConf("Message", conf), nil)
}

// MessageInfo is a component that displays a message with info color.
func MessageInfo(c *tgframe.Container, text string, conf ...*MessageConf) {
	color := tcutil.ColorInfo
	message(c, text, tgframe.OneConf("MessageInfo", conf), &color)
}

// MessageSuccess is a component that displays a message with success color.
func MessageSuccess(c *tgframe.Container, text string, conf ...*MessageConf) {
	color := tcutil.ColorSuccess
	message(c, text, tgframe.OneConf("MessageSuccess", conf), &color)
}

// MessageWarning is a component that displays a message with warning color.
func MessageWarning(c *tgframe.Container, text string, conf ...*MessageConf) {
	color := tcutil.ColorWarning
	message(c, text, tgframe.OneConf("MessageWarning", conf), &color)
}

// MessageDanger is a component that displays a message with danger color.
func MessageDanger(c *tgframe.Container, text string, conf ...*MessageConf) {
	color := tcutil.ColorDanger
	message(c, text, tgframe.OneConf("MessageDanger", conf), &color)
}

// message adds the message component. color is what the colored variant is
// named after, and overrides whatever the conf says; [Message] passes nil and
// leaves the conf's own color alone. The conf itself is not written to: the
// caller owns it and may reuse it across runs.
func message(c *tgframe.Container, text string, conf *MessageConf, color *tcutil.Color) {
	comp := newMessageComponent(text)
	comp.Color = conf.Color
	if color != nil {
		comp.Color = *color
	}

	comp.Title = conf.Title
	tgframe.SetConfID(comp, conf)

	c.AddComponent(comp)
}
