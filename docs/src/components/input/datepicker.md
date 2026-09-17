# Datepicker

Datepicker create a datepicker and return its selected date.

## API

```go
func Datepicker(c *tgframe.Container, label string, conf ...*DatepickerConf) *time.Time
```

* `c` is Parent container.
* `label` is the label for datepicker.
* `conf` is an optional configuration, at most one.
* Return the selected date, as midnight UTC on the day picked. nil if no date
  is selected.

```go
// DatepickerConf is the configuration for the Datepicker component.
type DatepickerConf struct {
	tgframe.Base // ID

	// Default is the date the picker starts on, read to the day. It is only
	// read until the app user first picks one.
	Default *time.Time

	// Disabled is true if the datepicker is disabled.
	Disabled bool
}

func (c *DatepickerConf) SetDefault(v time.Time) *DatepickerConf
```

Only the day is kept: a clock on the `Default` is dropped, the way
[Timepicker](timepicker.md) drops the date. That is what lets the three pickers
be compared and added without a conversion in between.

## Example

```go
dateValue := tgcomp.Datepicker(p.Main, "Datepicker")
if dateValue != nil {
	tgcomp.Text(p.Main, "Value: "+dateValue.Format("2006-01-02"),
		&tgcomp.TextConf{ID: "datepicker_result"})
}
```

Starting on a date, until the app user picks another:

```go
tgcomp.Datepicker(p.Main, "Datepicker",
	(&tgcomp.DatepickerConf{}).SetDefault(
		time.Date(2026, 9, 8, 0, 0, 0, 0, time.UTC)))
```

Clearing the picker is an answer of "no date": the return goes back to nil and
stays there, rather than falling back to `Default`.

![datepicker component](datepicker.png)
