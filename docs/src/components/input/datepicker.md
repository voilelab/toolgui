# DatePicker

DatePicker create a datepicker and return its selected date.

## API

```go
func DatePicker(c *tgframe.Container, label string, conf ...*DatePickerConf) *time.Time
```

* `c` is Parent container.
* `label` is the label for datepicker.
* `conf` is an optional configuration, at most one.
* Return the selected date, as midnight UTC on the day picked. nil if no date
  is selected.

```go
// DatePickerConf is the configuration for the DatePicker component.
type DatePickerConf struct {
	tgframe.Base // ID

	// Default is the date the picker starts on, read to the day. It is only
	// read until the app user first picks one.
	Default *time.Time

	// Disabled is true if the datepicker is disabled.
	Disabled bool
}

func (c *DatePickerConf) SetDefault(v time.Time) *DatePickerConf
```

Only the day is kept: a clock on the `Default` is dropped, the way
[TimePicker](timepicker.md) drops the date. That is what lets the three pickers
be compared and added without a conversion in between.

## Example

```go
{{#include ../../../demos/datepicker.go:demo}}
```

Starting on a date, until the app user picks another:

```go
tgcomp.DatePicker(p.Main, "DatePicker",
	(&tgcomp.DatePickerConf{}).SetDefault(
		time.Date(2026, 9, 8, 0, 0, 0, 0, time.UTC)))
```

Clearing the picker is an answer of "no date": the return goes back to nil and
stays there, rather than falling back to `Default`.

<div data-toolgui-demo="datepicker" data-toolgui-demo-height="440">

![datepicker component](datepicker.png)

</div>
