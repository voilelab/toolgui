# Menu

An App can declare a menu. The web frontend draws it as a menubar along the
top of the app, above the [side nav](sidenav.md) and the page.

```go
app.SetMenu(tgframe.NewMenu().
	Submenu("File", func(m *tgframe.Menu) {
		m.Text("Open", "file_open")
		m.Separator()
		m.Text("Quit", "file_quit")
	}).
	Submenu("Help", func(m *tgframe.Menu) {
		m.Text("About", "about")
	}))
```

A node is one of three things:

* `Text(label, id)`: an item that reports a click under `id`.
* `Separator()`: a line between the items around it.
* `Submenu(label, build)`: an item holding whatever `build` writes into it.
  Submenus nest.

All three work at the top of the tree as well as inside a submenu: a top level
`Text` is a button in the menubar, and a top level `Separator` divides the row.

An App that calls no `SetMenu` has no menubar, and the row is not in the
document at all. The menubar is also dropped in embed mode, along with the
rest of the app's chrome.

## Reading a click

A click on a menu item runs the current page, the same as a
[Button](../components/input/button.md) does, and the run handling it reports
the click:

```go
app.AddPage("index", "Index", func(p *tgframe.Params) error {
	if tgframe.MenuClicked(p, "file_open") {
		open()
	}

	return nil
})
```

Only that one run sees it, so a page that wants a pick to outlive the run has
to keep it — in the [State](../architecture/state-storage.md), usually.

A click on an id the menu does not declare is not a click: the id comes from
the client, and the menu is what says which ones exist.

## Why the menu belongs to the App

A menu is declared once on the App rather than drawn by a page func, unlike
every component. A page func runs again on every event, and a native menu has
no diff to apply: rebuilding the tree per run would rebuild the menu under the
user's pointer while they type. The tree being static is what lets the same Go
declaration stand for a menubar in the browser and for a real menu on the
desktop.

The same reason is why a menu item has no checkbox or radio form yet. A text
item only reports a click; a checkbox item would write state, under a key
shared with the page's components, and which namespace that key belongs to is
not settled.

## Menu ids and component ids

A menu item's click id and a component's live in one space, and an item is in
no run, so nothing can catch a collision between the two the way a run catches
two components sharing an id. Menu ids are kept apart by a reserved prefix
instead: the id `file_open` reports as `menu_item_file_open`, which no
component id can be, because a component's is `<component name>_<label>` and
every component name ends in `_component`.

`MenuClicked` adds the prefix itself. A page comparing against
[`State.GetClickID`](https://pkg.go.dev/github.com/voilelab/toolgui/toolgui/tgframe#State.GetClickID)
by hand wants `tgframe.MenuID("file_open")`.

Within the menu, an item with no label, a text item with no id, and two items
sharing one are all reported: `SetMenu` panics on them, so the app says so at
startup rather than serving a menubar quietly missing an item.
