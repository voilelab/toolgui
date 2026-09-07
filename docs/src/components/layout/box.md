# Box

Box provide a simple container that show box style.

## Usage

Box create a box container.

```go
func Box(c *tgframe.Container, conf ...*BoxConf) *tgframe.Container
```

* `c`: Parent container.
* `conf`: Optional configuration, at most one.

```go
// BoxConf is the configuration for the Box component.
type BoxConf struct {
	tgframe.Base // ID
}
```

The container a box hands out derives its id from the box's; give none and it
carries none, and the components inside are still placed by position.

## Example

```go
box := tgcomp.Box(boxCompCol)
tgcomp.Text(box, "A box!")
```

```go
box := tgcomp.Box(boxCompCol, &tgcomp.BoxConf{ID: "summary"})
```
