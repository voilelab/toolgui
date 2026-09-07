# Subtitle

Subtitle component display a subtitle.

The text supports [emoji shortcodes](emoji.md): `:tada:` renders as 🎉.

## API

```go
func Subtitle(c *tgframe.Container, text string, conf ...*SubtitleConf)
```

* `c` is Parent container.
* `text` is the subtitle text.
* `conf` is an optional configuration, at most one.

```go
// SubtitleConf is the configuration for the Subtitle component.
type SubtitleConf struct {
	tgframe.Base // ID
}
```

## Example

```go
tgcomp.Subtitle(p.Main, "Subtitle")
```

![subtitle component](subtitle.png)
