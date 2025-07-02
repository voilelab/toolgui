package tcinput

import (
	"fmt"
	"time"

	"github.com/voilelab/toolgui/toolgui/tgcomp/tcutil"
	"github.com/voilelab/toolgui/toolgui/tgframe"
)

var _ tgframe.Component = &datepickerComponent{}
var datepickerComponentName = "datepicker_component"

type datepickerComponent struct {
	*tgframe.BaseComponent
	Label string `json:"label"`
	Type  string `json:"type"`
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

// Datepicker create a datepicker and return its selected date.
// Return nil if no date is selected.
func Datepicker(s *tgframe.State, c *tgframe.Container, label string) *Date {
	comp := newDatepickerComponent(label, "date")
	c.AddComponent(comp)

	dateStr := s.GetString(comp.ID)
	if dateStr == nil {
		return nil
	}

	date, err := time.Parse("2006-01-02", *dateStr)
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

// Timepicker create a timepicker and return its selected time.
// Return nil if no time is selected.
func Timepicker(s *tgframe.State, c *tgframe.Container, label string) *Time {
	comp := newDatepickerComponent(label, "time")
	c.AddComponent(comp)

	timeStr := s.GetString(comp.ID)
	if timeStr == nil {
		return nil
	}

	t, err := time.Parse("15:04", *timeStr)
	if err != nil {
		panic(fmt.Sprintf("failed to parse time: %v", err))
	}

	return &Time{
		Hour: t.Hour(),
		Min:  t.Minute(),
	}
}

// Datetimepicker create a datetimepicker and return its selected datetime.
// Return nil if no datetime is selected.
func Datetimepicker(s *tgframe.State, c *tgframe.Container, label string) *time.Time {
	comp := newDatepickerComponent(label, "datetime-local")
	c.AddComponent(comp)

	datetimeStr := s.GetString(comp.ID)
	if datetimeStr == nil {
		return nil
	}

	datetime, err := time.Parse("2006-01-02T15:04", *datetimeStr)
	if err != nil {
		panic(fmt.Sprintf("failed to parse datetime: %v", err))
	}

	return &datetime
}
