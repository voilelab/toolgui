# Latex

Latex component is used to display LaTeX content.

## API

```go
func Latex(c *tgframe.Container, text string)
func LatexWithID(c *tgframe.Container, text string, id string)
```

* `c` is the container to add the LaTeX component to.
* `text` is the LaTeX content to display.
* `id` is a user specific element id.

## Example

```go
tgcomp.Latex(c, "E = mc^2")
```

