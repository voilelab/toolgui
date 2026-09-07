# Caption

Caption component displays a small dimmed text, for a note next to what it
explains.

The text supports [emoji shortcodes](emoji.md).

## API

```go
func Caption(c *tgframe.Container, text string, conf ...*CaptionConf)
```

* `c` is Parent container.
* `text` is the caption text.

`CaptionConf`:

| Field | Description                 | Default |
| ----- | --------------------------- | ------- |
| `ID`  | A user specific element id. | derived |

## Example

```go
tgcomp.Table(p.Main, []string{"a", "b"}, [][]string{{"1", "2"}})
tgcomp.Caption(p.Main, "Sampled hourly, UTC.")
```

![caption component](caption.png)
