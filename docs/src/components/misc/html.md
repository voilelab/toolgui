# Html

Html component is used to display html content.

## API

```go
func Html(c *tgframe.Container, html string, conf ...*HtmlConf)
```

### Parameters

* `c`: Parent container.
* `html`: Html content to display.
* `conf`: Optional configuration, at most one.

```go
// HtmlConf is the configuration for the Html component.
type HtmlConf struct {
	tgframe.Base // ID
}
```

## Example

```go
tgcomp.Html(p.Main, "<b>Hello world gen by html component</b>")
```
