# Box

Box provide a simple container that show box style.

## Usage

Box create a box container.

```go
func Box(c *tgframe.Container, id string) *tgframe.Container
```

* `c`: Parent container.
* `id`: Unique component ID.

## Example

```go
box := tgcomp.Box(boxCompCol, "box")
tgcomp.Text(box, "A box!")
```
