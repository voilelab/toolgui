# Popover

Popover puts a button on the page and hangs a floating panel behind it, so
options that are not needed often stay out of the way without taking a row of
their own.

## API

```go
func Popover(c *tgframe.Container, label string, conf ...*PopoverConf) *tgframe.Container
```

* `c` is the container to add the popover to.
* `label` is the label of the button that opens it.
* `conf` is an optional configuration, at most one.

```go
// PopoverConf is the configuration for the Popover component.
type PopoverConf struct {
	tgframe.Base // ID

	// Disabled is true if the trigger button is disabled.
	Disabled bool
}
```

The client keeps whether a popover is open, so it claims an id derived from
its label. Two popovers with the same label collide, and the run reports
`duplicated component id`; give one of them a `Conf.ID`.

## Example

```go
{{#include ../../../demos/popover.go:demo}}
```

```go
pop := tgcomp.Popover(c, "Advanced options",
	&tgcomp.PopoverConf{ID: "second_advanced", Disabled: true})
```

## Rendering

The panel holds whatever was written into the returned container, whether the
popover is open or not: nothing is built lazily, so a widget inside keeps what
it holds while the popover is closed.

The open state lives on the client, which the server never sees and never
sets. A widget inside the panel can rerun the page — the popover stays open
across the rerun. Clicking outside it, or pressing `Escape`, closes it.

`Escape` goes to whichever overlay was opened last, so a popover opened inside
a [dialog](dialog.md) closes on the first press and leaves the dialog for the
second.

An open panel follows its button: it hides itself while the button is scrolled
off the screen, and comes back when the button does.
