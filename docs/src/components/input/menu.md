# Menu

Menu puts a button on the page and hangs a list of actions behind it, and
returns the index of the item clicked. 0-indexed, `nil` if no item is clicked.

It is a button that carries several presses rather than one, so a row of
related actions — rename, duplicate, delete — takes a single button's worth of
space.

## API

```go
func Menu(c *tgframe.Container, label string, items []string, conf ...*MenuConf) *int
```

* `c` is the container to add the menu to.
* `label` is the text on the button that opens it.
* `items` are the actions in the dropdown, in the order they are shown.
* `conf` is an optional configuration, at most one.

```go
// MenuConf is the configuration for the Menu component.
type MenuConf struct {
	tgframe.Base // ID

	// Color defines the color of the button that opens the menu.
	Color tcutil.Color

	// Disabled is true if the button that opens the menu is disabled.
	Disabled bool
}
```

## Example

```go
{{#include ../../../demos/menu.go:demo}}
```

<div data-toolgui-demo="menu"></div>

## The click lasts one run

A menu reports a pick the way [`Button`](button.md) reports a press: the index
is there for exactly the run that handles the click, and `nil` again on the
next one. Act on it where you get it, or keep it yourself:

```go
if picked := tgcomp.Menu(p.Main, "Actions", items); picked != nil {
	applyAction(items[*picked])
}
```

There is no "currently selected item" to read back later. A menu is a list of
things to *do*; a list of things to *be* is [`Select`](select.md) or
[`Radio`](radio.md), which keep what was chosen.

## Identity

Each item claims a click id of its own, the menu's id with the item's index
behind it — `menu_component_Actions_0` for the first item of a menu labelled
`Actions`. Two menus with the same label therefore collide, and the run reports
`duplicated component id`; give one of them a `Conf.ID`, and its items follow:

```go
tgcomp.Menu(p.Main, "Actions", items)
tgcomp.Menu(p.Main, "Actions", items, &tgcomp.MenuConf{ID: "row_actions"})
```

Because the ids carry the index, a click naming an index the menu no longer has
— items dropped between the draw and the press — is not a pick, and the run
reads it as `nil` rather than as the last item.

## Rendering

Whether the dropdown is open lives on the client, which the server never sees
and never sets. The items are written every run, open or not.

Clicking an item closes the dropdown, as does clicking outside it or pressing
`Escape`. `Escape` goes to whichever overlay was opened last, so a menu opened
inside a [dialog](../layout/dialog.md) closes on the first press and leaves the
dialog for the second.

`Disabled` is on the button, not the items: a disabled menu cannot be opened at
all.

## Inside a form

Picking an item sends the [form](form.md) it is written in, the way pressing a
`Button` does. A menu item is an action to take now, so it does not sit in the
form's queue waiting for something else to send it: the run that reports the
pick is the one that reads the values queued before it.
