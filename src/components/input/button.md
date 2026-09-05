# Button

Button create a button and return true if it's clicked.

## API

```go
func Button(s *tgframe.State, c *tgframe.Container, label string) bool
func ButtonWithConf(s *tgframe.State, c *tgframe.Container, label string, conf *ButtonConf) bool
```

* `s` is State.
* `c` is Parent container.
* `label` is the text on button.
* `conf` is the configuration of the button.

```go
// ButtonConf is the configuration for the Button component
type ButtonConf struct {
	// Color defines the color of the button
	Color tcutil.Color

	// Disabled indicates whether the button should be initially disabled
	Disabled bool

	// ID is the unique identifier for this button component
	ID string
}
```

## Example

```go
btnClicked := tgcomp.Button(p.State, p.Main, "button")
if btnClicked {
	tgcomp.TextWithID(p.Main, "Value: true", "button_result")
} else {
	tgcomp.TextWithID(p.Main, "Value: false", "button_result")
}
```

![button component](button.png)
