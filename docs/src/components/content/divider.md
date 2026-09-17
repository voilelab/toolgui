# Divider

Divider component display a horizontal line.

## API

```go
func Divider(c *tgframe.Container, conf ...*DividerConf)
```

* `c` is Parent container.
* `conf` is an optional configuration, at most one.

```go
// DividerConf is the configuration for the Divider component.
type DividerConf struct {
	tgframe.Base // ID
}
```

## Example

```go
{{#include ../../../demos/divider.go:demo}}
```

![divider component](divider.png)
