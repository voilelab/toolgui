# Expand

Expand provides expandable layout.

## API

```go
func Expand(c *tgframe.Container, title string, expanded bool) *tgframe.Container
```

* `c` is the container to add the expandable component to.
* `title` is the title of the expandable component.
* `expanded` is the initial expanded state of the expandable component.

## Example

```go
tgcomp.Expand(c, "Expand", true)
```
