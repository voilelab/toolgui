# Radio

Radio create a group of radio items and return its selected value.

## API

```go
func Radio(c *tgframe.Container, label string, items []string, conf ...*RadioConf) *int
```

* `c` is Parent container.
* `label` is the label for options group.
* `items` is the list of options.
* `conf` is an optional configuration, at most one.
* Return the index of the selected item, 0-indexed. nil if no item is selected.

```go
// RadioConf is the configuration for the Radio component.
type RadioConf struct {
	tgframe.Base // ID

	// Default is the item the group starts on, as an index into items,
	// 0-based like the return. It is only read until the app user first
	// touches the component, and one that points outside items is ignored.
	Default *int

	// Disabled is true if the radio group is disabled.
	Disabled bool
}
```

## Example

```go
radioIndex := tgcomp.Radio(p.Main, "Radio", []string{"Value3", "Value4"})
if radioIndex != nil {
	tgcomp.Text(p.Main, fmt.Sprintf("Value: Value%d", *radioIndex+3),
		&tgcomp.TextConf{ID: "radio_result"})
}
```

Starting on the second item, until the app user picks another:

```go
second := 1
tgcomp.Radio(p.Main, "Radio", []string{"Value3", "Value4"},
	&tgcomp.RadioConf{Default: &second})
```

![radio component](radio.png)
