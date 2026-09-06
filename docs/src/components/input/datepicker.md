# Datepicker

Datepicker create a datepicker and return its selected date.

## API

```go
type Date struct {
	Year  int
	Month int
	Day   int
}

func Datepicker(s *tgframe.State, c *tgframe.Container, label string) *Date
```

* `s` is State.
* `c` is Parent container.
* `label` is the label for datepicker.
* Return the selected date. nil if no date is selected.

## Example

```go
dateValue := tgcomp.Datepicker(p.State, p.Main, "Datepicker")
if dateValue != nil {
	text := fmt.Sprintf("Value: %04d-%02d-%02d", dateValue.Year, dateValue.Month, dateValue.Day)
	tgcomp.TextWithID(p.Main, text, "datepicker_result")
}
```

![datepicker component](datepicker.png)
