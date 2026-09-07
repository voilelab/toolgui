# Select

Select create a select dropdown list and return its selected value.

## API

```go
func Select(c *tgframe.Container, label string, items []string, conf ...*SelectConf) *int
```

* `c` is Parent container.
* `label` is the label for select.
* `items` is the list of options.
* `conf` is an optional configuration, at most one.
* Return the index of the selected item, 0-indexed. nil if no item is selected.

```go
// SelectConf is the configuration for the Select component.
type SelectConf struct {
	tgframe.Base // ID
}
```

## Example

```go
selIndex := tgcomp.Select(p.Main, "Select", []string{"Value1", "Value2"})
if selIndex != nil {
	tgcomp.Text(p.Main, fmt.Sprintf("Value: Value%d", *selIndex+1),
		&tgcomp.TextConf{ID: "select_result"})
}
```

A select derives its id from its label, so two with the same label collide.
Naming either of them is the way out:

```go
tgcomp.Select(p.Main, "Pick", items)
tgcomp.Select(p.Main, "Pick", items, &tgcomp.SelectConf{ID: "second_pick"})
```

![select component](select.png)
