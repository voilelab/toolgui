# Text

Text component display a text.

The text supports [emoji shortcodes](emoji.md): `:tada:` renders as 🎉.

## API

```go
func Text(c *tgframe.Container, text string, conf ...*TextConf)
```

* `c` is Parent container.
* `text` is the text.
* `conf` is an optional configuration, at most one.

```go
// TextConf is the configuration for the Text component.
type TextConf struct {
	tgframe.Base // ID
}
```

## Example

```go
{{#include ../../../demos/text.go:demo}}
```

```go
tgcomp.Text(p.Main, "Text", &tgcomp.TextConf{ID: "greeting"})
```

Components are placed by position rather than by identity, so the same text
written twice shows twice:

```go
{{#include ../../../demos/text.go:duplicate}}
```

<div data-toolgui-demo="text">

![text component](text.png)

</div>
