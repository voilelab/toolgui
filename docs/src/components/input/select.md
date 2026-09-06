# Select

Select create a select dropdown list and return its selected value.

## API

```go
func Select(s *tgframe.State, c *tgframe.Container, label string, items []string) *int
```

* `s` is State.
* `c` is Parent container.
* `label` is the label for select.
* `items` is the list of options.
* Return the index of the selected item, 0-indexed. nil if no item is selected.

## Example

```go
selIndex := tgcomp.Select(p.State, p.Main, "Select", []string{"Value1", "Value2"})
if selIndex != nil {
	tgcomp.TextWithID(p.Main, fmt.Sprintf("Value: Value%d", *selIndex + 1), "select_result")
}
```

![select component](select.png)
