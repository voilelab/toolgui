# Timepicker

Timepicker create a timepicker and return its selected time.

## API

```go
type Time struct {
	Hour int
	Min  int
}

func Timepicker(c *tgframe.Container, label string, conf ...*TimepickerConf) *Time
```

* `c` is Parent container.
* `label` is the label for timepicker.
* `conf` is an optional configuration, at most one.
* Return the selected time. nil if no time is selected.

```go
// TimepickerConf is the configuration for the Timepicker component.
type TimepickerConf struct {
	tgframe.Base // ID

	// Default is the time the picker starts on. It is only read until the app
	// user first picks one, and a time that does not exist is ignored.
	Default *Time

	// Disabled is true if the timepicker is disabled.
	Disabled bool
}

func (c *TimepickerConf) SetDefault(v Time) *TimepickerConf
```

## Example

```go
timeValue := tgcomp.Timepicker(p.Main, "Timepicker")
if timeValue != nil {
	text := fmt.Sprintf("Value: %02d:%02d", timeValue.Hour, timeValue.Min)
	tgcomp.Text(p.Main, text, &tgcomp.TextConf{ID: "timepicker_result"})
}
```

Starting on a time, until the app user picks another:

```go
tgcomp.Timepicker(p.Main, "Timepicker",
	(&tgcomp.TimepickerConf{}).SetDefault(tgcomp.Time{Hour: 9, Min: 30}))
```

![timepicker component](timepicker.png)
