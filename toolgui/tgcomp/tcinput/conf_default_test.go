package tcinput_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/voilelab/toolgui/toolgui/tgcomp/tcinput"
	"github.com/voilelab/toolgui/toolgui/tgframe"
)

func ptr[T any](v T) *T {
	return &v
}

// componentProps renders the one component a call added the way it would be
// sent, which is where a Default has to show up for the widget on screen to
// start on it — the state alone only settles what Go reads back.
func componentProps(t *testing.T, packs []tgframe.NotifyPack) map[string]any {
	t.Helper()

	if len(packs) != 1 {
		t.Fatalf("got %d packs, want 1", len(packs))
	}

	bs, err := json.Marshal(packs[0])
	if err != nil {
		t.Fatal(err)
	}

	var got struct {
		Component map[string]any `json:"component"`
	}
	if err := json.Unmarshal(bs, &got); err != nil {
		t.Fatal(err)
	}

	return got.Component
}

// touch applies the event the frontend sends, so a test says "the app user
// answered" the same way a real session does.
func touch(state *tgframe.State, event tgframe.Event) *tgframe.State {
	event.ApplyState(state)
	return state
}

var fruits = []string{"apple", "banana", "cherry"}

func selectValue(state *tgframe.State, conf ...*tcinput.SelectConf) *int {
	return tcinput.Select(defaultContainer(state), "Fruit", fruits, conf...)
}

func TestSelectDefault(t *testing.T) {
	t.Run("untouched", func(t *testing.T) {
		got := selectValue(tgframe.NewState(), &tcinput.SelectConf{Default: ptr(1)})
		if got == nil || *got != 1 {
			t.Fatalf("Select = %v, want 1", got)
		}
	})

	// The regression the 1-based state used to cost the caller: index 0 is a
	// real item, not the "nothing selected" the state keeps 0 for.
	t.Run("first item", func(t *testing.T) {
		got := selectValue(tgframe.NewState(), &tcinput.SelectConf{Default: ptr(0)})
		if got == nil || *got != 0 {
			t.Fatalf("Select = %v, want 0", got)
		}
	})

	// Same index, on the wire, where it is 1-based. The caller never sees it.
	t.Run("first item is sent 1-based", func(t *testing.T) {
		var packs []tgframe.NotifyPack
		tcinput.Select(testContainer(tgframe.NewState(), &packs), "Fruit", fruits,
			&tcinput.SelectConf{Default: ptr(0)})

		if got := componentProps(t, packs)["default"]; got != 1.0 {
			t.Fatalf("default = %v, want 1", got)
		}
	})

	t.Run("picked beats default", func(t *testing.T) {
		state := touch(tgframe.NewState(), &tgframe.EventSelect{
			ID: "select_component_Fruit", Value: 3,
		})

		got := selectValue(state, &tcinput.SelectConf{Default: ptr(0)})
		if got == nil || *got != 2 {
			t.Fatalf("Select = %v, want 2", got)
		}
	})

	// Clearing the select is an answer too, so the default does not come back.
	t.Run("cleared beats default", func(t *testing.T) {
		state := touch(tgframe.NewState(), &tgframe.EventSelect{
			ID: "select_component_Fruit", Value: 0,
		})

		if got := selectValue(state, &tcinput.SelectConf{Default: ptr(1)}); got != nil {
			t.Fatalf("Select = %v, want nil", *got)
		}
	})

	t.Run("out of range is ignored", func(t *testing.T) {
		for _, def := range []int{-1, 3} {
			var packs []tgframe.NotifyPack
			got := tcinput.Select(testContainer(tgframe.NewState(), &packs),
				"Fruit", fruits, &tcinput.SelectConf{Default: ptr(def)})

			if got != nil {
				t.Errorf("Select with default %d = %v, want nil", def, *got)
			}
			if sent := componentProps(t, packs)["default"]; sent != 0.0 {
				t.Errorf("default sent for %d = %v, want 0", def, sent)
			}
		}
	})

	t.Run("no default", func(t *testing.T) {
		if got := selectValue(tgframe.NewState()); got != nil {
			t.Fatalf("Select = %v, want nil", *got)
		}
	})

	// A selection made when the list was longer, which is a stale index the
	// caller would otherwise index items with.
	t.Run("stale selection is dropped", func(t *testing.T) {
		state := tgframe.NewState()
		state.Set("select_component_Fruit", 9)

		if got := selectValue(state); got != nil {
			t.Fatalf("Select = %v, want nil", *got)
		}
	})

	t.Run("disabled", func(t *testing.T) {
		var packs []tgframe.NotifyPack
		tcinput.Select(testContainer(tgframe.NewState(), &packs), "Fruit", fruits,
			&tcinput.SelectConf{Disabled: true})

		if got := componentProps(t, packs)["disabled"]; got != true {
			t.Fatalf("disabled = %v, want true", got)
		}
	})
}

func radioValue(state *tgframe.State, conf ...*tcinput.RadioConf) *int {
	return tcinput.Radio(defaultContainer(state), "Fruit", fruits, conf...)
}

func TestRadioDefault(t *testing.T) {
	t.Run("untouched", func(t *testing.T) {
		got := radioValue(tgframe.NewState(), &tcinput.RadioConf{Default: ptr(1)})
		if got == nil || *got != 1 {
			t.Fatalf("Radio = %v, want 1", got)
		}
	})

	// A radio has no placeholder to reserve an index for, so 0 stays 0 on the
	// wire — the asymmetry with [tcinput.Select] the state doc warns about.
	t.Run("first item is sent 0-based", func(t *testing.T) {
		var packs []tgframe.NotifyPack
		got := tcinput.Radio(testContainer(tgframe.NewState(), &packs), "Fruit", fruits,
			&tcinput.RadioConf{Default: ptr(0)})

		if got == nil || *got != 0 {
			t.Fatalf("Radio = %v, want 0", got)
		}
		if sent := componentProps(t, packs)["default"]; sent != 0.0 {
			t.Fatalf("default = %v, want 0", sent)
		}
	})

	t.Run("picked beats default", func(t *testing.T) {
		state := touch(tgframe.NewState(), &tgframe.EventSelect{
			ID: "radio_component_Fruit", Value: 0,
		})

		got := radioValue(state, &tcinput.RadioConf{Default: ptr(2)})
		if got == nil || *got != 0 {
			t.Fatalf("Radio = %v, want 0", got)
		}
	})

	t.Run("out of range is ignored", func(t *testing.T) {
		for _, def := range []int{-1, 3} {
			var packs []tgframe.NotifyPack
			got := tcinput.Radio(testContainer(tgframe.NewState(), &packs),
				"Fruit", fruits, &tcinput.RadioConf{Default: ptr(def)})

			if got != nil {
				t.Errorf("Radio with default %d = %v, want nil", def, *got)
			}
			if sent := componentProps(t, packs)["default"]; sent != nil {
				t.Errorf("default sent for %d = %v, want null", def, sent)
			}
		}
	})

	t.Run("no default", func(t *testing.T) {
		if got := radioValue(tgframe.NewState()); got != nil {
			t.Fatalf("Radio = %v, want nil", *got)
		}
	})

	t.Run("stale selection is dropped", func(t *testing.T) {
		state := tgframe.NewState()
		state.Set("radio_component_Fruit", 9)

		if got := radioValue(state); got != nil {
			t.Fatalf("Radio = %v, want nil", *got)
		}
	})

	t.Run("disabled", func(t *testing.T) {
		var packs []tgframe.NotifyPack
		tcinput.Radio(testContainer(tgframe.NewState(), &packs), "Fruit", fruits,
			&tcinput.RadioConf{Disabled: true})

		if got := componentProps(t, packs)["disabled"]; got != true {
			t.Fatalf("disabled = %v, want true", got)
		}
	})
}

func TestDatepickerDefault(t *testing.T) {
	const id = "datepicker_component_Day"

	conf := func() *tcinput.DatepickerConf {
		return &tcinput.DatepickerConf{
			Default: &tcinput.Date{Year: 2026, Month: 9, Day: 8},
		}
	}

	t.Run("untouched", func(t *testing.T) {
		got := tcinput.Datepicker(defaultContainer(tgframe.NewState()), "Day", conf())
		if got == nil || got.String() != "2026-09-08" {
			t.Fatalf("Datepicker = %v, want 2026-09-08", got)
		}
	})

	t.Run("sent to the frontend", func(t *testing.T) {
		var packs []tgframe.NotifyPack
		tcinput.Datepicker(testContainer(tgframe.NewState(), &packs), "Day", conf())

		if got := componentProps(t, packs)["default"]; got != "2026-09-08" {
			t.Fatalf("default = %v, want 2026-09-08", got)
		}
	})

	t.Run("picked beats default", func(t *testing.T) {
		state := touch(tgframe.NewState(), &tgframe.EventInput{
			ID: id, Value: "2020-01-02",
		})

		got := tcinput.Datepicker(defaultContainer(state), "Day", conf())
		if got == nil || got.String() != "2020-01-02" {
			t.Fatalf("Datepicker = %v, want 2020-01-02", got)
		}
	})

	// Clearing the picker used to panic on the empty string it lands; it is
	// an answer of "no date", and the default does not come back.
	t.Run("cleared beats default", func(t *testing.T) {
		state := touch(tgframe.NewState(), &tgframe.EventInput{ID: id, Value: ""})

		if got := tcinput.Datepicker(defaultContainer(state), "Day", conf()); got != nil {
			t.Fatalf("Datepicker = %v, want nil", got)
		}
	})

	t.Run("impossible date is ignored", func(t *testing.T) {
		for _, def := range []*tcinput.Date{
			{Year: 2026, Month: 13, Day: 1},
			{Year: 2026, Month: 2, Day: 30},
		} {
			var packs []tgframe.NotifyPack
			got := tcinput.Datepicker(testContainer(tgframe.NewState(), &packs), "Day",
				&tcinput.DatepickerConf{Default: def})

			if got != nil {
				t.Errorf("Datepicker with default %v = %v, want nil", def, got)
			}
			if sent := componentProps(t, packs)["default"]; sent != "" {
				t.Errorf("default sent for %v = %q, want empty", def, sent)
			}
		}
	})

	t.Run("disabled", func(t *testing.T) {
		var packs []tgframe.NotifyPack
		tcinput.Datepicker(testContainer(tgframe.NewState(), &packs), "Day",
			&tcinput.DatepickerConf{Disabled: true})

		if got := componentProps(t, packs)["disabled"]; got != true {
			t.Fatalf("disabled = %v, want true", got)
		}
	})
}

func TestTimepickerDefault(t *testing.T) {
	const id = "datepicker_component_At"

	conf := func() *tcinput.TimepickerConf {
		return &tcinput.TimepickerConf{Default: &tcinput.Time{Hour: 9, Min: 30}}
	}

	t.Run("untouched", func(t *testing.T) {
		got := tcinput.Timepicker(defaultContainer(tgframe.NewState()), "At", conf())
		if got == nil || got.String() != "09:30" {
			t.Fatalf("Timepicker = %v, want 09:30", got)
		}
	})

	t.Run("sent to the frontend", func(t *testing.T) {
		var packs []tgframe.NotifyPack
		tcinput.Timepicker(testContainer(tgframe.NewState(), &packs), "At", conf())

		if got := componentProps(t, packs)["default"]; got != "09:30" {
			t.Fatalf("default = %v, want 09:30", got)
		}
	})

	t.Run("picked beats default", func(t *testing.T) {
		state := touch(tgframe.NewState(), &tgframe.EventInput{ID: id, Value: "18:45"})

		got := tcinput.Timepicker(defaultContainer(state), "At", conf())
		if got == nil || got.String() != "18:45" {
			t.Fatalf("Timepicker = %v, want 18:45", got)
		}
	})

	t.Run("cleared beats default", func(t *testing.T) {
		state := touch(tgframe.NewState(), &tgframe.EventInput{ID: id, Value: ""})

		if got := tcinput.Timepicker(defaultContainer(state), "At", conf()); got != nil {
			t.Fatalf("Timepicker = %v, want nil", got)
		}
	})

	t.Run("impossible time is ignored", func(t *testing.T) {
		got := tcinput.Timepicker(defaultContainer(tgframe.NewState()), "At",
			&tcinput.TimepickerConf{Default: &tcinput.Time{Hour: 24, Min: 0}})

		if got != nil {
			t.Fatalf("Timepicker = %v, want nil", got)
		}
	})

	t.Run("disabled", func(t *testing.T) {
		var packs []tgframe.NotifyPack
		tcinput.Timepicker(testContainer(tgframe.NewState(), &packs), "At",
			&tcinput.TimepickerConf{Disabled: true})

		if got := componentProps(t, packs)["disabled"]; got != true {
			t.Fatalf("disabled = %v, want true", got)
		}
	})
}

func TestDatetimepickerDefault(t *testing.T) {
	const id = "datepicker_component_When"

	conf := func() *tcinput.DatetimepickerConf {
		return &tcinput.DatetimepickerConf{
			Default: ptr(time.Date(2026, 9, 8, 13, 5, 0, 0, time.UTC)),
		}
	}

	t.Run("untouched", func(t *testing.T) {
		got := tcinput.Datetimepicker(defaultContainer(tgframe.NewState()), "When", conf())
		if got == nil || got.Format("2006-01-02T15:04") != "2026-09-08T13:05" {
			t.Fatalf("Datetimepicker = %v, want 2026-09-08T13:05", got)
		}
	})

	t.Run("sent to the frontend", func(t *testing.T) {
		var packs []tgframe.NotifyPack
		tcinput.Datetimepicker(testContainer(tgframe.NewState(), &packs), "When", conf())

		if got := componentProps(t, packs)["default"]; got != "2026-09-08T13:05" {
			t.Fatalf("default = %v, want 2026-09-08T13:05", got)
		}
	})

	// The wire carries minutes, so a default with seconds on it reads back
	// the same way a pick of that minute would.
	t.Run("read to the minute", func(t *testing.T) {
		got := tcinput.Datetimepicker(defaultContainer(tgframe.NewState()), "When",
			&tcinput.DatetimepickerConf{
				Default: ptr(time.Date(2026, 9, 8, 13, 5, 59, 0, time.UTC)),
			})

		if got == nil || !got.Equal(time.Date(2026, 9, 8, 13, 5, 0, 0, time.UTC)) {
			t.Fatalf("Datetimepicker = %v, want 2026-09-08 13:05:00 UTC", got)
		}
	})

	t.Run("picked beats default", func(t *testing.T) {
		state := touch(tgframe.NewState(), &tgframe.EventInput{
			ID: id, Value: "2020-01-02T07:00",
		})

		got := tcinput.Datetimepicker(defaultContainer(state), "When", conf())
		if got == nil || got.Format("2006-01-02T15:04") != "2020-01-02T07:00" {
			t.Fatalf("Datetimepicker = %v, want 2020-01-02T07:00", got)
		}
	})

	t.Run("cleared beats default", func(t *testing.T) {
		state := touch(tgframe.NewState(), &tgframe.EventInput{ID: id, Value: ""})

		if got := tcinput.Datetimepicker(defaultContainer(state), "When", conf()); got != nil {
			t.Fatalf("Datetimepicker = %v, want nil", got)
		}
	})

	t.Run("disabled", func(t *testing.T) {
		var packs []tgframe.NotifyPack
		tcinput.Datetimepicker(testContainer(tgframe.NewState(), &packs), "When",
			&tcinput.DatetimepickerConf{Disabled: true})

		if got := componentProps(t, packs)["disabled"]; got != true {
			t.Fatalf("disabled = %v, want true", got)
		}
	})
}

// TestSetDefaultBuilders pins the shorthand the pointer Defaults need to be
// writable where the conf is, the way NumberConf.SetDefault already is.
func TestSetDefaultBuilders(t *testing.T) {
	t.Run("select", func(t *testing.T) {
		got := selectValue(tgframe.NewState(), (&tcinput.SelectConf{}).SetDefault(2))
		if got == nil || *got != 2 {
			t.Fatalf("Select = %v, want 2", got)
		}
	})

	t.Run("radio", func(t *testing.T) {
		got := radioValue(tgframe.NewState(), (&tcinput.RadioConf{}).SetDefault(2))
		if got == nil || *got != 2 {
			t.Fatalf("Radio = %v, want 2", got)
		}
	})

	t.Run("datepicker", func(t *testing.T) {
		got := tcinput.Datepicker(defaultContainer(tgframe.NewState()), "Day",
			(&tcinput.DatepickerConf{}).SetDefault(
				tcinput.Date{Year: 2026, Month: 9, Day: 8}))

		if got == nil || got.String() != "2026-09-08" {
			t.Fatalf("Datepicker = %v, want 2026-09-08", got)
		}
	})

	t.Run("timepicker", func(t *testing.T) {
		got := tcinput.Timepicker(defaultContainer(tgframe.NewState()), "At",
			(&tcinput.TimepickerConf{}).SetDefault(tcinput.Time{Hour: 9, Min: 30}))

		if got == nil || got.String() != "09:30" {
			t.Fatalf("Timepicker = %v, want 09:30", got)
		}
	})

	t.Run("datetimepicker", func(t *testing.T) {
		want := time.Date(2026, 9, 8, 13, 5, 0, 0, time.UTC)
		got := tcinput.Datetimepicker(defaultContainer(tgframe.NewState()), "When",
			(&tcinput.DatetimepickerConf{}).SetDefault(want))

		if got == nil || !got.Equal(want) {
			t.Fatalf("Datetimepicker = %v, want %v", got, want)
		}
	})
}

// A fileupload has no Default to test — see [tcinput.FileuploadConf] for why —
// so Disabled is the whole of its share of this.
func TestFileuploadDisabled(t *testing.T) {
	var packs []tgframe.NotifyPack
	tcinput.Fileupload(testContainer(tgframe.NewState(), &packs), "Doc", ".pdf",
		&tcinput.FileuploadConf{Disabled: true})

	if got := componentProps(t, packs)["disabled"]; got != true {
		t.Fatalf("disabled = %v, want true", got)
	}
}
