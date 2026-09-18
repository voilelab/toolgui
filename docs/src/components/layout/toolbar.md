# Toolbar

Toolbar is a row of controls: what is written into it lines up horizontally,
each item as wide as it needs to be, rather than taking a row of the page
each. `Sticky` keeps that row at the top of the page while the rest of it
scrolls under.

## Usage

```go
func Toolbar(c *tgframe.Container, conf ...*ToolbarConf) *tgframe.Container
```

* `c`: Parent container.
* `conf`: Optional configuration, at most one.

```go
// ToolbarConf is the configuration for the Toolbar component.
type ToolbarConf struct {
	tgframe.Base // ID

	// Sticky keeps the row at the top of the page while the rest of it
	// scrolls under, on an opaque background so nothing shows through.
	Sticky bool

	// Justify is ToolbarJustifyStart (default), ToolbarJustifyEnd or
	// ToolbarJustifyBetween. Anything else panics.
	Justify string
}
```

The container a toolbar hands out derives its id from the toolbar's; give none
and it carries none, and the components inside are still placed by position.

`Justify` says where the row's spare width goes:

| Value | Items sit |
| --- | --- |
| `ToolbarJustifyStart` | at the start of the row, the default |
| `ToolbarJustifyEnd` | at the end of the row |
| `ToolbarJustifyBetween` | spread out, first and last at the edges |

A row too wide for the viewport wraps rather than pushing the page sideways,
so a toolbar is as safe on a phone as a [column](column.md) is.

## Example

```go
{{#include ../../../demos/toolbar.go:demo}}
```

<div data-toolgui-demo="toolbar" data-toolgui-demo-height="480"></div>

## Sticky

A sticky toolbar stays at the top of the viewport while the page scrolls under
it, on the page's own background and with a line under it, so what passes
beneath does not show through:

```go
{{#include ../../../demos/toolbar.go:sticky}}
```

It sticks to whatever scrolls the page, which for an app in the shell is the
document. A toolbar written inside something that scrolls on its own — a
[dialog](dialog.md) body, say — sticks to the top of that instead.

## Toolbar or Column?

[`Column`](column.md) is for laying a page out in parts: each column is a
share of the width, and what goes in one is a section of the page. A toolbar
is for the controls above that page — it packs them together at their natural
widths, wraps them on a narrow screen and, with `Sticky`, keeps them reachable
however far down the page the reader is.
