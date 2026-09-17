# DateTimePicker

DateTimePicker create a datetimepicker and return its selected datetime.

## API

```go
func DateTimePicker(c *tgframe.Container, label string, conf ...*DateTimePickerConf) *time.Time
```

* `c` is Parent container.
* `label` is the label for datetimepicker.
* `conf` is an optional configuration, at most one.
* Return the selected datetime. nil if no datetime is selected.

```go
// DateTimePickerConf is the configuration for the DateTimePicker component.
type DateTimePickerConf struct {
	tgframe.Base // ID

	// Default is the datetime the picker starts on, read to the minute. It is
	// only read until the app user first picks one.
	Default *time.Time

	// Disabled is true if the datetimepicker is disabled.
	Disabled bool
}

func (c *DateTimePickerConf) SetDefault(v time.Time) *DateTimePickerConf
```

## Example

```go
dateValue := tgcomp.DateTimePicker(p.Main, "DateTimePicker")
if dateValue != nil {
	tgcomp.Text(p.Main, "Value: "+dateValue.Format("2006-01-02 15:04"),
		&tgcomp.TextConf{ID: "datetimepicker_result"})
}
```

Starting on a datetime, until the app user picks another:

```go
tgcomp.DateTimePicker(p.Main, "DateTimePicker",
	(&tgcomp.DateTimePickerConf{}).SetDefault(
		time.Date(2026, 9, 8, 13, 5, 0, 0, time.UTC)))
```

The wire carries minutes, so a `Default` with seconds on it is read to the
minute — the same value a pick of that minute gives.

![datetimepicker component](datetimepicker.png)
