# Html

Html component is used to display html content.

## API

```go
func Html(c *tgframe.Container, html string)
func HtmlWithID(c *tgframe.Container, html string, id string)
```

### Parameters

* `c`: Parent container.
* `html`: Html content to display.
* `id`: Id of the component.

## Example

```go
tgcomp.Html(p.Main, "<b>Hello world gen by html component</b>")
```
