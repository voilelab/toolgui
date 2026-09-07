# Badge

Badge component displays a short label, for a status or a tag next to other
content.

The text supports [emoji shortcodes](emoji.md).

## API

```go
func Badge(c *tgframe.Container, text string, conf ...*BadgeConf)
```

* `c` is Parent container.
* `text` is the badge text.

`BadgeConf`:

| Field   | Description                                                       | Default          |
| ------- | ----------------------------------------------------------------- | ---------------- |
| `Color` | `tcutil.ColorInfo`, `ColorSuccess`, `ColorWarning` or `ColorDanger`. | neutral        |
| `ID`    | A user specific element id.                                        | derived          |

## Example

```go
tgcomp.Badge(p.Main, "Badge")
tgcomp.Badge(p.Main, "Shipped", &tgcomp.BadgeConf{Color: tcutil.ColorSuccess})
```

![badge component](badge.png)
