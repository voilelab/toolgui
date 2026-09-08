# Toggle

Toggle create a switch and return true if it's on.

It is [Checkbox](checkbox.md) drawn as a switch: same signature, same value,
different affordance. Reach for it where the setting takes effect immediately
("dark mode on"), and for a checkbox where it is one of several things being
filled in before a submit.

## API

### Interface

```go
func Toggle(c *tgframe.Container, label string, conf ...*ToggleConf) bool
```

### Parameters

* `c` is Parent container.
* `label` is the text beside the switch.
* `conf` is an optional configuration, at most one.

```go
// ToggleConf is the configuration for a toggle.
type ToggleConf struct {
	tgframe.Base // ID

	// Default is true if the toggle is default on.
	Default bool

	// Disabled is true if the toggle is disabled.
	Disabled bool
}
```

## Example

```go
toggleValue := tgcomp.Toggle(p.Main, "Toggle")
tgcomp.Text(p.Main, fmt.Sprint("Value: ", toggleValue),
	&tgcomp.TextConf{ID: "toggle_result"})
```
