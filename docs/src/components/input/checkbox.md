# Checkbox

Checkbox create a checkbox and return true if it's checked.

## API

### Interface

```go
func Checkbox(c *tgframe.Container, label string, conf ...*CheckboxConf) bool
```

### Parameters

* `c` is Parent container.
* `label` is the text on checkbox.
* `conf` is an optional configuration, at most one.

```go
// CheckboxConf is the configuration for a checkbox.
type CheckboxConf struct {
	tgframe.Base // ID

	// Default is true if the checkbox is default checked.
	Default bool

	// Disabled is true if the checkbox is disabled.
	Disabled bool
}
```

## Example

```go
checkboxValue := tgcomp.Checkbox(p.Main, "Checkbox")
tgcomp.Text(p.Main, fmt.Sprint("Value: ", checkboxValue),
	&tgcomp.TextConf{ID: "checkbox_result"})
```

![checkbox component](checkbox.png)
