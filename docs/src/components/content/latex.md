# Latex

Latex component is used to display LaTeX content.

## API

```go
func Latex(c *tgframe.Container, text string, conf ...*LatexConf)
```

* `c` is the container to add the LaTeX component to.
* `text` is the LaTeX content to display.
* `conf` is an optional configuration, at most one.

```go
// LatexConf is the configuration for the Latex component.
type LatexConf struct {
	tgframe.Base // ID
}
```

## Example

```go
tgcomp.Latex(c, "E = mc^2")
```
