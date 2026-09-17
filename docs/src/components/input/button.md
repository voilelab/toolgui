# Button

Button create a button and return true if it's clicked.

## API

```go
func Button(c *tgframe.Container, label string, conf ...*ButtonConf) bool
```

* `c` is Parent container.
* `label` is the text on button.
* `conf` is an optional configuration, at most one.

```go
// ButtonConf is the configuration for the Button component
type ButtonConf struct {
	tgframe.Base // ID

	// Color defines the color of the button
	Color tcutil.Color

	// Disabled indicates whether the button should be initially disabled
	Disabled bool
}
```

## Example

```go
{{#include ../../../demos/button.go:demo}}
```

Two buttons with the same label claim the same id, so one of them needs its
own:

```go
tgcomp.Button(p.Main, "Save")
tgcomp.Button(p.Main, "Save", &tgcomp.ButtonConf{ID: "save_all"})
```

![button component](button.png)

## Asking before the button is drawn

```go
func ButtonClicked(c *tgframe.Container, label string, conf ...*ButtonConf) bool
```

`Button` reports the click where it is written, which is a problem when the
button belongs *below* the content it changes — a "Load details" at the bottom
of a card. `ButtonClicked` reads the click off the run's state instead, so the
page can handle it first and write the content once:

```go
if tgcomp.ButtonClicked(p.Main, "Load details") {
	details = fetchDetails()
}

tgcomp.Text(p.Main, details)
tgcomp.Button(p.Main, "Load details")
```

It reads the click, it does not draw the button: give it the same `label` and
`conf` the `Button` call gets, and it derives the same id. Where the button
carries a `ButtonConf.ID`, pass that conf to both — sharing one variable is the
way to keep them from drifting apart:

```go
conf := &tgcomp.ButtonConf{ID: "load_details"}

if tgcomp.ButtonClicked(p.Main, "Load details", conf) {
	details = fetchDetails()
}

tgcomp.Text(p.Main, details)
tgcomp.Button(p.Main, "Load details", conf)
```

It is true for exactly the run that handles the click on a button the page has
on the screen — the same as what `Button` returns there — and false again on
the next run. A click id naming a button the last run never drew is not one:
the click comes from the client, and `ButtonClicked` is asked before the page
has written anything to check it against, so it checks that itself.

Don't compare `State.GetClickID()` with the conf id yourself: the click id
carries the component name in front of it
([Identity](../../architecture/components.md#the-name-in-front-of-the-id)),
and comparing it by hand skips that check.
