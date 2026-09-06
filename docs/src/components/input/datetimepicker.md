# Datetimepicker

Datetimepicker create a datetimepicker and return its selected datetime.

## API

```go
func Datetimepicker(s *tgframe.State, c *tgframe.Container, label string) *time.Time
```

* `s` is State.
* `c` is Parent container.
* `label` is the label for datetimepicker.
* Return the selected datetime. nil if no datetime is selected.
## Example

```go
dateValue := tgcomp.Datetimepicker(p.State, p.Main, "Datetimepicker")
if dateValue != nil {
	tgcomp.TextWithID(p.Main, "Value: "+dateValue.Format("2006-01-02 15:04"), "datetimepicker_result")
}
```

![datetimepicker component](datetimepicker.png)
