# Markdown

Prose supports [emoji shortcodes](emoji.md). A shortcode in a code span or a
fenced block stays as written.

Formulas are written in LaTeX and drawn by KaTeX, as in the
[Latex](latex.md) component: `$...$` inline, `$$...$$` as a block. An inline
formula can't start or end with a space, so `$5 and $10` stays text.

## API

### Interface

```go
func Markdown(c *tgframe.Container, markdown string, conf ...*MarkdownConf)
```

### Parameters

* `c`: Parent container.
* `markdown`: Markdown content to display.
* `conf`: Optional configuration, at most one.

```go
// MarkdownConf is the configuration for the Markdown component.
type MarkdownConf struct {
	tgframe.Base // ID
}
```

## Example

```go
tgcomp.Markdown(c, "* Hello, World!")
```

```go
tgcomp.Markdown(c, "Euler: $e^{i\\pi} + 1 = 0$")
```

```go
tgcomp.Markdown(c, "* Hello, World!", &tgcomp.MarkdownConf{ID: "my_markdown"})
```
