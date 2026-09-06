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
	// Title is the title of the message. Optional.
	Title string

	// Color is the color of the message. Default is tcutil.ColorNull.
	Color tcutil.Color

	// ID is the unique identifier of the component.
	ID string
}

// Message is a component that displays a message.
func Message(c *tgframe.Container, text string) {
	MessageWithConf(c, text, nil)
}

// MessageInfo is a component that displays a message with info color.
func MessageInfo(c *tgframe.Container, text string) {
	MessageWithConf(c, text, &MessageConf{Color: tcutil.ColorInfo})
}

// MessageSuccess is a component that displays a message with success color.
func MessageSuccess(c *tgframe.Container, text string) {
	MessageWithConf(c, text, &MessageConf{Color: tcutil.ColorSuccess})
}

// MessageWarning is a component that displays a message with warning color.
func MessageWarning(c *tgframe.Container, text string) {
	MessageWithConf(c, text, &MessageConf{Color: tcutil.ColorWarning})
}

// MessageDanger is a component that displays a message with danger color.
func MessageDanger(c *tgframe.Container, text string) {
	MessageWithConf(c, text, &MessageConf{Color: tcutil.ColorDanger})
}

// MessageWithConf is a component that displays a message with extra config.
func MessageWithConf(c *tgframe.Container, text string, conf *MessageConf) {
	if conf == nil {
		conf = &MessageConf{}
	}

	comp := newMessageComponent(text)
	comp.Color = conf.Color
	comp.Title = conf.Title

	if conf.ID != "" {
		comp.SetID(conf.ID)
	}

	c.AddComponent(comp)
}
