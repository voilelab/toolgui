package tcinput

import (
	"fmt"
	"time"

	"github.com/voilelab/toolgui/toolgui/tgcomp/tcutil"
	"github.com/voilelab/toolgui/toolgui/tgframe"
)

var _ tgframe.Component = &datepickerComponent{}
var datepickerComponentName = "datepicker_component"

// The formats the three pickers put on the wire, and read back off it.
const (
	dateFormat     = "2006-01-02"
	timeFormat     = "15:04"
	datetimeFormat = "2006-01-02T15:04"
)

type datepickerComponent struct {
	*tgframe.BaseComponent
	Label string `json:"label"`
	Type  string `json:"type"`

	// Default is the value in the picker's own wire format, empty for none.
	Default  string `json:"default"`
	Disabled bool   `json:"disabled"`
}

func newDatepickerComponent(label string, typ string) *datepickerComponent {
	return &datepickerComponent{
		BaseComponent: &tgframe.BaseComponent{
			Name: datepickerComponentName,
			ID:   tcutil.NormalID(datepickerComponentName, label),
		},
		Label: label,
		Type:  typ,
	}
}

// Date is the selected date.
type Date struct {
	// Year is the selected year. Format: 2006
	Year int

	// Month is the selected month. Format: 1-12
	Month int

	// Day is the selected day. Format: 1-31
	Day int
}

func (d *Date) String() string {
	return fmt.Sprintf("%04d-%02d-%02d", d.Year, d.Month, d.Day)
}

// normalizeDate drops a default that is not a date anyone could have picked —
// a thirteenth month, a thirtieth of February — so a bad one is ignored rather
// than reaching the frontend as a string it cannot read. The copy keeps the
// caller's Default out of the return.
func normalizeDate(d *Date) *Date {
	if d == nil {
		return nil
	}

	t, err := time.Parse(dateFormat, d.String())
	if err != nil {
		return nil
	}

	return &Date{Year: t.Year(), Month: int(t.Month()), Day: t.Day()}
}

// DatepickerConf is the configuration for the Datepicker component.
type DatepickerConf struct {
	tgframe.Base

	// Default is the date the picker starts on. It is only read until the app
	// user first picks one, and a date that does not exist is ignored.
	Default *Date

	// Disabled is true if the datepicker is disabled.
	Disabled bool
}

// SetDefault sets Default from a value, so a literal can be written where the
// conf is.
func (c *DatepickerConf) SetDefault(v Date) *DatepickerConf {
	c.Default = &v
	return c
}

// Datepicker create a datepicker and return its selected date.
// Return nil if no date is selected.
func Datepicker(c *tgframe.Container, label string, conf ...*DatepickerConf) *Date {
	cf := tgframe.OneConf("Datepicker", conf)

	comp := newDatepickerComponent(label, "date")
	def := normalizeDate(cf.Default)
	if def != nil {
		comp.Default = def.String()
	}
	comp.Disabled = cf.Disabled
	tgframe.SetConfID(comp, cf)
	c.AddComponent(comp)

	dateStr := c.State.GetString(comp.ID)
	if dateStr == nil {
		return def
	}

	// The app user cleared the picker: a selection of nothing, not a value to
	// parse, and not a reason to put the default back.
	if *dateStr == "" {
		return nil
	}

	date, err := time.Parse(dateFormat, *dateStr)
	if err != nil {
		panic(fmt.Sprintf("failed to parse date: %v", err))
	}

	return &Date{
		Year:  date.Year(),
		Month: int(date.Month()),
		Day:   date.Day(),
	}
}

// Time is the selected time.
type Time struct {
	// Hour is the selected hour. Format: 0-23
	Hour int

	// Min is the selected minute. Format: 0-59
	Min int
}

func (t *Time) String() string {
	return fmt.Sprintf("%02d:%02d", t.Hour, t.Min)
}

// normalizeTime is [normalizeDate] for a time of day.
func normalizeTime(t *Time) *Time {
	if t == nil {
		return nil
	}

	parsed, err := time.Parse(timeFormat, t.String())
	if err != nil {
		return nil
	}

	return &Time{Hour: parsed.Hour(), Min: parsed.Minute()}
}

// TimepickerConf is the configuration for the Timepicker component.
type TimepickerConf struct {
	tgframe.Base

	// Default is the time the picker starts on. It is only read until the app
	// user first picks one, and a time that does not exist is ignored.
	Default *Time

	// Disabled is true if the timepicker is disabled.
	Disabled bool
}

// SetDefault sets Default from a value, so a literal can be written where the
// conf is.
func (c *TimepickerConf) SetDefault(v Time) *TimepickerConf {
	c.Default = &v
	return c
}

// Timepicker create a timepicker and return its selected time.
// Return nil if no time is selected.
func Timepicker(c *tgframe.Container, label string, conf ...*TimepickerConf) *Time {
	cf := tgframe.OneConf("Timepicker", conf)

	comp := newDatepickerComponent(label, "time")
	def := normalizeTime(cf.Default)
	if def != nil {
		comp.Default = def.String()
	}
	comp.Disabled = cf.Disabled
	tgframe.SetConfID(comp, cf)
	c.AddComponent(comp)

	timeStr := c.State.GetString(comp.ID)
	if timeStr == nil {
		return def
	}

	if *timeStr == "" {
		return nil
	}

	t, err := time.Parse(timeFormat, *timeStr)
	if err != nil {
		panic(fmt.Sprintf("failed to parse time: %v", err))
	}

	return &Time{
		Hour: t.Hour(),
		Min:  t.Minute(),
	}
}

// normalizeDatetime rounds a default down to the minute the wire carries, and
// into the UTC a value read back off it lands in, so a default and a pick of
// the same moment are the same value.
func normalizeDatetime(t *time.Time) *time.Time {
	if t == nil {
		return nil
	}

	out, err := time.Parse(datetimeFormat, t.Format(datetimeFormat))
	if err != nil {
		return nil
	}

	return &out
}

// DatetimepickerConf is the configuration for the Datetimepicker component.
type DatetimepickerConf struct {
	tgframe.Base

	// Default is the datetime the picker starts on, read to the minute. It is
	// only read until the app user first picks one.
	Default *time.Time

	// Disabled is true if the datetimepicker is disabled.
	Disabled bool
}

// SetDefault sets Default from a value, so a literal can be written where the
// conf is.
func (c *DatetimepickerConf) SetDefault(v time.Time) *DatetimepickerConf {
	c.Default = &v
	return c
}

// Datetimepicker create a datetimepicker and return its selected datetime.
// Return nil if no datetime is selected.
func Datetimepicker(c *tgframe.Container, label string, conf ...*DatetimepickerConf) *time.Time {
	cf := tgframe.OneConf("Datetimepicker", conf)

	comp := newDatepickerComponent(label, "datetime-local")
	def := normalizeDatetime(cf.Default)
	if def != nil {
		comp.Default = def.Format(datetimeFormat)
	}
	comp.Disabled = cf.Disabled
	tgframe.SetConfID(comp, cf)
	c.AddComponent(comp)

	datetimeStr := c.State.GetString(comp.ID)
	if datetimeStr == nil {
		return def
	}

	if *datetimeStr == "" {
		return nil
	}

	datetime, err := time.Parse(datetimeFormat, *datetimeStr)
	if err != nil {
		panic(fmt.Sprintf("failed to parse datetime: %v", err))
	}

	return &datetime
}
