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
btnClicked := tgcomp.Button(p.Main, "button")
tgcomp.Text(p.Main, fmt.Sprint("Value: ", btnClicked),
	&tgcomp.TextConf{ID: "button_result"})
```

Two buttons with the same label claim the same id, so one of them needs its
own:

```go
tgcomp.Button(p.Main, "Save")
tgcomp.Button(p.Main, "Save", &tgcomp.ButtonConf{ID: "save_all"})
```

![button component](button.png)
