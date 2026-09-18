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

* `Text(label, id)`: an item that reports a click under `id`. It takes an
  optional [`MenuTextConf`](#accelerators).
* `Separator()`: a line between the items around it.
* `Submenu(label, build)`: an item holding whatever `build` writes into it.
  Submenus nest.

All three work at the top of the tree as well as inside a submenu: a top level
`Text` is a button in the menubar, and a top level `Separator` divides the row.

An App that calls no `SetMenu` has no menubar, and the row is not in the
document at all. The menubar is also dropped in embed mode, along with the
rest of the app's chrome.

The row is armed by a click: crossing the menubar opens nothing until an entry
has been clicked, and from then on moving along the row moves the open dropdown
with the pointer. Only one entry is open at a time.

The row scrolls away with the page rather than pinning to the top of the
viewport, which would cover the first line of whatever is under it.

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

## Accelerators

An item can declare a key combination that fires it without the menu being
opened:

```go
m.Text("Open", "file_open", &tgframe.MenuTextConf{
	Accelerator: "CmdOrCtrl+O",
})
```

One declaration, written once, in a spelling that belongs to no platform:

* Modifiers are `CmdOrCtrl`, `OptionOrAlt`, `Shift` and `Ctrl`, joined to the
  key by `+`. `CmdOrCtrl` is Command on macOS and Control everywhere else, and
  `OptionOrAlt` is Option on macOS and Alt everywhere else — which is what
  saves the app from writing the combination twice. `Ctrl` is Control on every
  platform, so `CmdOrCtrl+Ctrl` is refused: off macOS the two are one key.
* The key is a letter, a digit, one of `` ` - = [ ] \ ; ' , . / ``, or one of
  `backspace`, `tab`, `enter`, `escape`, `left`, `right`, `up`, `down`,
  `space`, `delete`, `home`, `end`, `page up`, `page down`, `f1` to `f24`.
* Case does not matter, and neither does the order the modifiers are written
  in.

An accelerator names a *key*, not the character the key produces, so a
character that needs Shift is not one of them: `CmdOrCtrl+?` is refused, and
`CmdOrCtrl+Shift+/` is how to say it. This is also the only spelling the two
carriers agree on — a browser reports `?` for that keystroke while the desktop
menu is handed `/` — and it is why `+` is written `Shift+=` rather than being
the key that joins the parts.

The two carriers serve it differently, which is the whole of why it is worth
declaring rather than wiring up:

* On the desktop the combination hangs off the native menu item and the OS
  dispatches it. Nothing in the app listens for a keystroke.
* In a browser nothing dispatches it, so the shell listens on the document
  itself, matches the keystroke against the tree, and sends the item's click.
  A menu that declares no accelerator at all has no listener: the row is drawn
  and nothing is bound.

Either way the item reports the same click id, and the run handling it reads
it back with `MenuClicked` without knowing which of the two fired.

`SetMenu` panics on a combination it cannot serve, and on two items landing on
one keystroke. What counts as one keystroke is what is actually held down, not
how it was written: `CmdOrCtrl+O` and `cmdorctrl+o` are one, and so are
`CmdOrCtrl+O` and `Ctrl+O`, because `CmdOrCtrl` *is* Control off macOS and the
second item would never fire there. `CmdOrCtrl+O` and `CmdOrCtrl+Shift+O` are
two, wherever they run.

### Typing is not a shortcut

A combination carrying no modifier but `Shift` is ignored while the focus is
in a text field — an `input`, a `textarea`, a `select`, or anything
`contenteditable`. `F2` on a menu item is a shortcut everywhere on the page
except inside the box being filled in, where it is a keystroke.

A real chord — anything carrying `CmdOrCtrl`, `Ctrl` or `OptionOrAlt` — still
reaches the menu from inside a text field, the way `Cmd+S` does in an editor.

One chord is not a chord: a keystroke carrying AltGr is left alone. On Windows,
and on some layouts elsewhere, AltGr is reported as Control and Alt held
together, so `AltGr+E` and a `Ctrl+OptionOrAlt+E` accelerator arrive as the
same event. There is no telling them apart, so the character wins — firing the
item would swallow a keystroke the visitor meant to type. A `Ctrl+OptionOrAlt`
accelerator is therefore not reachable by keyboard on such a layout, which is a
reason to prefer `CmdOrCtrl` and `Shift` for one.

### What the browser has already taken

The browser sees a keystroke first. The shell calls `preventDefault` on every
combination it matched, which is as far as a page's say goes, and for some of
them it is not far enough: the browser takes the keystroke at a level no page
is asked about, and the item never hears it.

These are not reliably the app's, on at least one major browser:

| Combination | Who takes it |
| --- | --- |
| `CmdOrCtrl+T`, `CmdOrCtrl+N`, `CmdOrCtrl+Shift+N`, `CmdOrCtrl+Shift+T` | New tab or window, and reopening a closed one. Not cancellable. |
| `CmdOrCtrl+W`, `CmdOrCtrl+Q` | Closing the tab or quitting. Not cancellable. |
| `CmdOrCtrl+Tab`, `CmdOrCtrl+1` to `CmdOrCtrl+9` | Switching tabs. |
| `f1` to `f12` | The browser's own: `F1` help, `F3` find again, `F5` reload, `F6` address bar, `F11` full screen, `F12` devtools. `F2` and `F4` are usually free; the rest are not. |
| `escape` | Stops a load, and closes whatever is open on the page first. |
| `CmdOrCtrl+P`, `CmdOrCtrl+S`, `CmdOrCtrl+O`, `CmdOrCtrl+F` | Print, save, open, find. A page can cancel all four, but an extension or a user setting can take them back. |

The first three rows are the ones to stay off: nothing a page does reaches
them. The last two are a page's to take, and taking them means the browser's
own behaviour is gone for as long as the app is open — `CmdOrCtrl+F` on a menu
item costs the visitor the browser's find bar.

A combination is not rejected for being on this list. An app that only ever
runs on the desktop should say `CmdOrCtrl+O` and mean it; one that runs in
both places has to decide which behaviour it would rather have. Silently
dropping the declaration on the web would hide the choice rather than settle
it.

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

Within the menu, an item with no label, a text item with no id, two items
sharing one, an accelerator that cannot be parsed, and two items sharing one
of those are all reported: `SetMenu` panics on them, so the app says so at
startup rather than serving a menubar quietly missing an item.
