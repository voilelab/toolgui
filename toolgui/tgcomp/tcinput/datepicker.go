package tcinput

import (
	"time"

	"github.com/voilelab/toolgui/toolgui/tgcomp/tcutil"
	"github.com/voilelab/toolgui/toolgui/tgframe"
	"github.com/voilelab/toolgui/toolgui/tgutil"
)

var _ tgframe.Component = &datepickerComponent{}
var datepickerComponentName = "datepicker_component"

// picker is everything the three pickers differ by: the input type the
// frontend renders, the format that goes on the wire with it, and the noun an
// error names. The bodies are otherwise the same, and shared.
type picker struct {
	typ    string
	format string
	noun   string
}

var (
	datePicker     = picker{"date", "2006-01-02", "date"}
	timePicker     = picker{"time", "15:04", "time"}
	datetimePicker = picker{"datetime-local", "2006-01-02T15:04", "datetime"}
)

type datepickerComponent struct {
	*tgframe.BaseComponent
	Label string `json:"label"`
	Type  string `json:"type"`

	// Default is the value in the picker's own wire format, empty for none.
	Default  string `json:"default"`
	Disabled bool   `json:"disabled"`
}

func newDatePickerComponent(label string, typ string) *datepickerComponent {
	return &datepickerComponent{
		BaseComponent: &tgframe.BaseComponent{
			Name: datepickerComponentName,
			ID:   tcutil.NormalID(datepickerComponentName, label),
		},
		Label: label,
		Type:  typ,
	}
}

// normalize rounds a time to what this picker's format can carry, and into the
// UTC a value read back off the wire lands in, so a default and a pick of the
// same moment are the same time.Time. A time the format cannot carry at all is
// dropped, which keeps a bad default out of the frontend rather than sending it
// a string it cannot read.
//
// Each picker keeps only its own half: a date read back sits at midnight UTC, a
// time of day on 1 January year 0, the date time.Parse fills in for a layout
// that names none. That is what makes the three comparable — the part a picker
// does not ask for is left where the format puts it rather than invented. The
// half that is kept is read in the caller's own zone, so a Default just before
// midnight in +08:00 is the date it reads as there.
func (p picker) normalize(t *time.Time) *time.Time {
	if t == nil {
		return nil
	}

	out, err := time.Parse(p.format, t.Format(p.format))
	if err != nil {
		return nil
	}

	return &out
}

// DatePickerConf is the configuration for the DatePicker component.
type DatePickerConf struct {
	tgframe.Base

	// Default is the date the picker starts on, read to the day. It is only
	// read until the app user first picks one.
	Default *time.Time

	// Disabled is true if the datepicker is disabled.
	Disabled bool
}

// SetDefault sets Default from a value, so a literal can be written where the
// conf is.
func (c *DatePickerConf) SetDefault(v time.Time) *DatePickerConf {
	c.Default = &v
	return c
}

// DatePicker create a datepicker and return its selected date, as midnight UTC
// on the day picked. Return nil if no date is selected.
//
// Only the day is kept: a clock the caller's Default carried is dropped, the
// way [TimePicker] drops the date, so the three pickers hand back times that
// compare.
func DatePicker(c *tgframe.Container, label string, conf ...*DatePickerConf) *time.Time {
	cf := tgframe.OneConf("DatePicker", conf)
	return datePicker.pick(c, label, cf.Default, cf.Disabled, cf)
}

// TimePickerConf is the configuration for the TimePicker component.
type TimePickerConf struct {
	tgframe.Base

	// Default is the time of day the picker starts on, read to the minute. It
	// is only read until the app user first picks one.
	Default *time.Time

	// Disabled is true if the timepicker is disabled.
	Disabled bool
}

// SetDefault sets Default from a value, so a literal can be written where the
// conf is.
func (c *TimePickerConf) SetDefault(v time.Time) *TimePickerConf {
	c.Default = &v
	return c
}

// TimePicker create a timepicker and return its selected time of day, as that
// clock on 1 January year 0 in UTC — only the clock is meaningful. Return nil
// if no time is selected.
//
// Only the clock is kept: a date the caller's Default carried is dropped, the
// way [DatePicker] drops the clock.
func TimePicker(c *tgframe.Container, label string, conf ...*TimePickerConf) *time.Time {
	cf := tgframe.OneConf("TimePicker", conf)
	return timePicker.pick(c, label, cf.Default, cf.Disabled, cf)
}

// DateTimePickerConf is the configuration for the DateTimePicker component.
type DateTimePickerConf struct {
	tgframe.Base

	// Default is the datetime the picker starts on, read to the minute. It is
	// only read until the app user first picks one.
	Default *time.Time

	// Disabled is true if the datetimepicker is disabled.
	Disabled bool
}

// SetDefault sets Default from a value, so a literal can be written where the
// conf is.
func (c *DateTimePickerConf) SetDefault(v time.Time) *DateTimePickerConf {
	c.Default = &v
	return c
}

// DateTimePicker create a datetimepicker and return its selected datetime,
// read to the minute and in UTC. Return nil if no datetime is selected.
func DateTimePicker(c *tgframe.Container, label string,
	conf ...*DateTimePickerConf) *time.Time {

	cf := tgframe.OneConf("DateTimePicker", conf)
	return datetimePicker.pick(c, label, cf.Default, cf.Disabled, cf)
}

// pick is the body the three pickers share.
func (p picker) pick(c *tgframe.Container, label string, def *time.Time,
	disabled bool, conf tgframe.Conf) *time.Time {

	comp := newDatePickerComponent(label, p.typ)
	def = p.normalize(def)
	if def != nil {
		comp.Default = def.Format(p.format)
	}
	comp.Disabled = disabled
	tgframe.SetConfID(comp, conf)
	c.AddComponent(comp)

	str := c.State.GetString(comp.ID)
	if str == nil {
		return def
	}

	// The app user cleared the picker: a selection of nothing, not a value to
	// parse, and not a reason to put the default back.
	if *str == "" {
		return nil
	}

	t, err := time.Parse(p.format, *str)
	if err != nil {
		c.Fail(tgutil.Errorf("failed to parse %s: %w", p.noun, err))
		return nil
	}

	return &t
}
