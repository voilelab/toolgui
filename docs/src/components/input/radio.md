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

![radio component](radio.png)
