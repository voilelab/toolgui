# Dialog

Dialog asks the app user something without leaving the page: a confirmation, a
short form. Its body is only computed while it is open, so the query behind
"delete these 3 rows?" does not run on every rerun.

## Usage

```go
func Dialog(c *tgframe.Container, title string, conf ...*DialogConf) *DialogContainer
```

* `c`: Parent container.
* `title`: The heading of the dialog.
* `conf`: Optional configuration, at most one.

```go
// DialogConf is the configuration for the Dialog component.
type DialogConf struct {
	tgframe.Base // ID

	// Width is DialogWidthSmall (default), DialogWidthMedium or
	// DialogWidthLarge. Anything else panics.
	Width string

	// Dismissible lets the app user close the dialog with X, ESC or a click
	// outside, default on. Set it with SetDismissible.
	Dismissible *bool
}
```

`Dismissible` is a pointer, against the positive naming the other confs use
for their bools, because the default is on: a plain bool's zero value would
mean "not dismissible", which is the rarer of the two. Set it through the
helper rather than taking the address of a variable:

```go
(&tgcomp.DialogConf{}).SetDismissible(false)
```

The dialog hands back a handle:

```go
// Open opens the dialog, from anywhere in the run.
func (d *DialogContainer) Open()

// Close closes the dialog.
func (d *DialogContainer) Close()

// IsOpen reports whether the dialog is open, as of this point in the run.
func (d *DialogContainer) IsOpen() bool

// With writes the body, and calls f only while the dialog is open.
func (d *DialogContainer) With(f func(c *tgframe.Container))
```

Whether the dialog is open is kept by the page rather than the client, so a
dialog claims an id derived from its title. Two dialogs with the same title
collide; give one of them a `Conf.ID`.

## Example

```go
d := tgcomp.Dialog(p.Main, "Delete confirm")

if tgcomp.Button(p.Main, "Delete") {
	d.Open()
}

d.With(func(c *tgframe.Container) {
	tgcomp.Text(c, "Delete "+name+"?")
	if tgcomp.Button(c, "Yes, delete") {
		remove(name)
		d.Close()
	}
})
```

## Handle the trigger first, then `With`

`With` is what draws the body, so whatever opens the dialog has to be handled
above it, as in the example. `Open()` still works below `With` — the dialog is
open on the client the moment it is called, wherever in the run that is,
because being open is a property of the dialog rather than "was the body
sent". But the body of *that* run is empty: `With` already went by while the
dialog was closed. It appears on the next run.

`Close()` is the mirror image. The body of the run it is called in has already
been sent, so the client is told to close separately and keeps what it holds
until the run ends. The next run does not write the body, and the client drops
it then. Nothing has to cut the run short.

## Widget state while closed

Widgets inside the body keep their state while the dialog is closed, the same
as any component hidden behind an `if`: a closed dialog does not write them,
and a page only gives an id's state back when something takes it off the
screen. So a half-filled form is still half-filled when the dialog is reopened.

To throw that state away instead, write the body into an
[`Empty`](empty.md) slot and clear it when the dialog closes: clearing a slot
gives back the ids it held, and the state under them goes with it.

## Rendering

The dialog is a portal: it covers the window from wherever it is written, so
one declared in `p.Sidebar` still darkens the whole page rather than the side
column. A closed dialog renders as an empty portal and takes up no room.

Two dialogs may be open at once, and they stack in the order the page writes
them rather than the order they were opened: the later one is drawn over the
earlier. A dialog opened from inside another is written after it, which is the
order that gives, so nothing special is needed for the usual case:

```go
d := tgcomp.Dialog(p.Main, "Delete confirm")
why := tgcomp.Dialog(p.Main, "What deleting does")

d.With(func(c *tgframe.Container) {
	if tgcomp.Button(c, "What does this do?") {
		why.Open()
	}
})

why.With(func(c *tgframe.Container) {
	tgcomp.Text(c, "The rows are removed for good.")
})
```

`why.Open()` is called from `d`'s body, which is above `why.With`, so the
second dialog's body is drawn in the same run it opens in.

ESC and a click outside reach the topmost dialog only, so dismissing the one
on top leaves the one under it open. A dialog with `Dismissible` off on top
swallows them rather than letting the dialog beneath take them.
