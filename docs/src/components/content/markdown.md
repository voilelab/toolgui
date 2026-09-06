# Markdown

## API

### Interface

```go
func Markdown(c *tgframe.Container, markdown string)
func MarkdownWithID(c *tgframe.Container, markdown string, id string)
```

### Parameters

* `c`: Parent container.
* `markdown`: Markdown content to display.
* `id`: Unique component ID.

## Example

```go
tgcomp.Markdown(c, "* Hello, World!")
```

```go
tgcomp.MarkdownWithID(c, "* Hello, World!", "my-markdown")
```
