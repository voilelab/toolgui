# Chat Input

ChatInput draws a box to send a chat message. It returns the message on the
run the send starts, and empties the box.

## API

### Interface

```go
func ChatInput(c *tgframe.Container, placeholder string, conf ...*ChatInputConf) (string, bool)
```

### Parameters

* `c`: Parent container.
* `placeholder`: The hint shown in the empty box. It also names the input, so
  two inputs with the same placeholder collide; give one a `Conf.ID`.
* `conf`: Optional configuration, at most one.

```go
// ChatInputConf is the configuration for the ChatInput component.
type ChatInputConf struct {
	tgframe.Base // ID

	// MaxLength is the most characters a message may have. 0 is no limit.
	MaxLength int

	// Disabled keeps the app user from sending, e.g. while a reply streams.
	Disabled bool
}
```

### Return

The message and `true` on the run its send starts; `""` and `false` on every
other run. The message is not kept: store it, as a chat keeps its history.

Enter sends, Shift+Enter breaks the line. An Enter that ends an IME
composition only picks the characters.

### Where it is drawn

Written in the page's main container, `p.Main`, it is pinned to the bottom of
the page wherever it is written. That lets a page read the message first and
draw the messages it adds after, as the example does. In any other container
it stays where it is written.

## Example

```go
{{#include ../../../demos/chat_input.go:turn}}
```

```go
{{#include ../../../demos/chat_input.go:demo}}
```

`echo` stands in for a model, streamed with
[WriteStream](../content/write_stream.md):

```go
{{#include ../../../demos/chat_input.go:echo}}
```

<div data-toolgui-demo="chat_input" data-toolgui-demo-height="520"></div>
