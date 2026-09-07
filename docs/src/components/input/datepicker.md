# Datepicker

Datepicker create a datepicker and return its selected date.

## API

```go
type Date struct {
	Year  int
	Month int
	Day   int
}

func Datepicker(c *tgframe.Container, label string, conf ...*DatepickerConf) *Date
```

* `c` is Parent container.
* `label` is the label for datepicker.
* `conf` is an optional configuration, at most one.
* Return the selected date. nil if no date is selected.

```go
// DatepickerConf is the configuration for the Datepicker component.
type DatepickerConf struct {
	tgframe.Base // ID
}
```

## Example

```go
dateValue := tgcomp.Datepicker(p.Main, "Datepicker")
if dateValue != nil {
	text := fmt.Sprintf("Value: %04d-%02d-%02d",
		dateValue.Year, dateValue.Month, dateValue.Day)
	tgcomp.Text(p.Main, text, &tgcomp.TextConf{ID: "datepicker_result"})
}
```

![datepicker component](datepicker.png)
