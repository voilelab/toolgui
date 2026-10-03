# Chat Message

ChatMessage draws one message of a chat, and returns the container its content
goes in. Any component can go in it.

## API

### Interface

```go
func ChatMessage(c *tgframe.Container, role string, conf ...*ChatMessageConf) *tgframe.Container
```

### Parameters

* `c`: Parent container.
* `role`: Who sent the message. `"user"` and `"human"` are drawn as the user,
  `"assistant"` and `"ai"` as the assistant, each with an icon of its own. Any
  other role is drawn by its first letter.
* `conf`: Optional configuration, at most one.

```go
// ChatMessageConf is the configuration for the ChatMessage component.
type ChatMessageConf struct {
	tgframe.Base // ID

	// Avatar is an emoji, or an image URL: http, https, data:image or a path
	// starting with / or ./. Default is an icon for the user and the
	// assistant, and the role's first letter for any other role.
	Avatar string
}
```

## Example

```go
{{#include ../../../demos/chat_message.go:demo}}
```

A reply streamed with [WriteStream](../content/write_stream.md):

```go
{{#include ../../../demos/chat_message.go:stream}}
```

<div data-toolgui-demo="chat_message" data-toolgui-demo-height="520"></div>
