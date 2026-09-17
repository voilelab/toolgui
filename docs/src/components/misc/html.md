# HTML

HTML component is used to display html content.

## API

```go
func HTML(c *tgframe.Container, html string, conf ...*HTMLConf)
```

### Parameters

* `c`: Parent container.
* `html`: HTML content to display.
* `conf`: Optional configuration, at most one.

```go
// HTMLConf is the configuration for the HTML component.
type HTMLConf struct {
	tgframe.Base // ID
}
```

## Example

```go
{{#include ../../../demos/html.go:demo}}
```
