# Message

Message is a component that displays a message.

## API

### Interface

```go
func Message(c *tgframe.Container, text string, conf ...*MessageConf)
func MessageInfo(c *tgframe.Container, text string, conf ...*MessageConf)
func MessageSuccess(c *tgframe.Container, text string, conf ...*MessageConf)
func MessageWarning(c *tgframe.Container, text string, conf ...*MessageConf)
func MessageDanger(c *tgframe.Container, text string, conf ...*MessageConf)
```

### Parameters

* `c`: Parent container.
* `text`: Text to display.
* `conf`: Optional configuration, at most one.

`Message[Info|Success|Warning|Danger]` set `Color` themselves and ignore what
the conf says; `Message` follows the conf.

```go
type MessageConf struct {
	tgframe.Base // ID

	// Title is the title of the message. Optional.
	Title string

	// Color is the color of the message. Default is tcutil.ColorNull.
	Color tcutil.Color
}
```

## Example

```go
tgcomp.Message(c, "Hello, World!")
```

```go
tgcomp.MessageInfo(c, "Hello, World!")
```

```go
tgcomp.Message(c, "Hello, World!", &tgcomp.MessageConf{
	Title: "Info",
	Color: tcutil.ColorInfo,
})
```

```go
tgcomp.MessageDanger(c, "It broke", &tgcomp.MessageConf{Title: "danger!"})
```
