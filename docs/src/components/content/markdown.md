# Markdown

Prose supports [emoji shortcodes](emoji.md). A shortcode in a code span or a
fenced block stays as written.

## API

### Interface

```go
func Markdown(c *tgframe.Container, markdown string, conf ...*MarkdownConf)
```

### Parameters

* `c`: Parent container.
* `markdown`: Markdown content to display.
* `conf`: Optional configuration, at most one.

```go
// MarkdownConf is the configuration for the Markdown component.
type MarkdownConf struct {
	tgframe.Base // ID
}
```

## Example

```go
tgcomp.Markdown(c, "* Hello, World!")
```

```go
tgcomp.Markdown(c, "* Hello, World!", &tgcomp.MarkdownConf{ID: "my_markdown"})
```
