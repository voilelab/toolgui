# Title

Title component display a title.

## API

```go
func Title(c *tgframe.Container, text string)
func TitleWithID(c *tgframe.Container, text string, id string)
```

* `c` is Parent container.
* `text` is the title text.
* `id` is a user specific element id.

## Example

```go
tgcomp.Title(p.Main, "Title")
```

![title component](title.png)
