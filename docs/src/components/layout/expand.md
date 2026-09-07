# Expand

Expand provides expandable layout.

## API

```go
func Expand(c *tgframe.Container, title string, expanded bool, conf ...*ExpandConf) *tgframe.Container
```

* `c` is the container to add the expandable component to.
* `title` is the title of the expandable component.
* `expanded` is the initial expanded state of the expandable component.
* `conf` is an optional configuration, at most one.

```go
// ExpandConf is the configuration for the Expand component.
type ExpandConf struct {
	tgframe.Base // ID
}
```

The client keeps whether an expander is open, so it claims an id derived from
its title. Two expanders with the same title collide; give one of them a
`Conf.ID`.

## Example

```go
tgcomp.Expand(c, "Expand", true)
```

```go
tgcomp.Expand(c, "Details", false, &tgcomp.ExpandConf{ID: "second_details"})
```

## Rendering

The contents are built on the first open and then stay rendered, hidden while
collapsed. An expander that was never opened costs nothing, and collapsing one
that was opened keeps what it holds.
