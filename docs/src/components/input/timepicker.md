# Timepicker

Timepicker create a timepicker and return its selected time of day.

## API

```go
func Timepicker(c *tgframe.Container, label string, conf ...*TimepickerConf) *time.Time
```

* `c` is Parent container.
* `label` is the label for timepicker.
* `conf` is an optional configuration, at most one.
* Return the selected time of day, as that clock on 1 January year 0 in UTC —
  only the clock is meaningful. nil if no time is selected.

```go
// TimepickerConf is the configuration for the Timepicker component.
type TimepickerConf struct {
	tgframe.Base // ID

	// Default is the time of day the picker starts on, read to the minute. It
	// is only read until the app user first picks one.
	Default *time.Time

	// Disabled is true if the timepicker is disabled.
	Disabled bool
}

func (c *TimepickerConf) SetDefault(v time.Time) *TimepickerConf
```

Only the clock is kept: a date on the `Default` is dropped, the way
[Datepicker](datepicker.md) drops the clock. Read the hour and minute off it
rather than the date:

```go
day := tgcomp.Datepicker(p.Main, "Day")
at := tgcomp.Timepicker(p.Main, "At")
if day != nil && at != nil {
	when := day.Add(time.Duration(at.Hour())*time.Hour +
		time.Duration(at.Minute())*time.Minute)
	tgcomp.Text(p.Main, "Value: "+when.Format("2006-01-02 15:04"))
}
```

## Example

```go
timeValue := tgcomp.Timepicker(p.Main, "Timepicker")
if timeValue != nil {
	tgcomp.Text(p.Main, "Value: "+timeValue.Format("15:04"),
		&tgcomp.TextConf{ID: "timepicker_result"})
}
```

Starting on a time, until the app user picks another:

```go
tgcomp.Timepicker(p.Main, "Timepicker",
	(&tgcomp.TimepickerConf{}).SetDefault(
		time.Date(0, time.January, 1, 9, 30, 0, 0, time.UTC)))
```

![timepicker component](timepicker.png)
