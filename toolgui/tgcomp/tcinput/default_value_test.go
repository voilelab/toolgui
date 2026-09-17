package tcinput_test

import (
	"reflect"
	"testing"
	"time"

	"github.com/voilelab/toolgui/toolgui/tgcomp/tcinput"
	"github.com/voilelab/toolgui/toolgui/tgframe"
)

func defaultContainer(state *tgframe.State) *tgframe.Container {
	return tgframe.NewContainer("test", state, func(tgframe.NotifyPack) {})
}

func ptr[T any](v T) *T { return &v }

// checkDefault runs one component down both of the paths Default has to hold
// on: an untouched state, where the component reads as its conf's Default, and
// the state the app user's input would have landed under key, where it reads
// as that input instead.
func checkDefault[T any](t *testing.T, name string, key string, input any,
	call func(*tgframe.Container) T, wantDefault, wantInput T) {

	t.Helper()

	t.Run(name, func(t *testing.T) {
		got := call(defaultContainer(tgframe.NewState()))
		if !reflect.DeepEqual(got, wantDefault) {
			t.Errorf("untouched = %v, want the conf's default %v",
				deref(got), deref(wantDefault))
		}

		state := tgframe.NewState()
		state.Set(key, input)

		got = call(defaultContainer(state))
		if !reflect.DeepEqual(got, wantInput) {
			t.Errorf("after input = %v, want %v", deref(got), deref(wantInput))
		}
	})
}

// deref is for the failure message only: half these components hand back a
// pointer, and the address in it says nothing about what went wrong.
func deref(v any) any {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Pointer || rv.IsNil() {
		return v
	}

	return rv.Elem().Interface()
}

// TestDefaultValue is the whole table at once: for every component that takes
// a Default, Default is what Go reads before the app user has touched it, and
// the app user's input is what Go reads after. One case per component, so a
// component whose Default stops reaching the return cannot hide behind its
// neighbours.
//
// [tcinput.Fileupload] is the one input with no Default — a file input cannot
// be filled in from script — so it has no case here.
func TestDefaultValue(t *testing.T) {
	fruits := []string{"apple", "banana", "cherry"}

	checkDefault(t, "textbox", "textbox_component_Name", "typed",
		func(c *tgframe.Container) string {
			return tcinput.Textbox(c, "Name", &tcinput.TextboxConf{Default: "toolgui"})
		}, "toolgui", "typed")

	checkDefault(t, "textarea", "textarea_component_Body", "typed",
		func(c *tgframe.Container) string {
			return tcinput.Textarea(c, "Body", &tcinput.TextareaConf{Default: "toolgui"})
		}, "toolgui", "typed")

	checkDefault(t, "checkbox", "checkbox_component_Agree", false,
		func(c *tgframe.Container) bool {
			return tcinput.Checkbox(c, "Agree", &tcinput.CheckboxConf{Default: true})
		}, true, false)

	checkDefault(t, "toggle", "toggle_component_Notify", false,
		func(c *tgframe.Container) bool {
			return tcinput.Toggle(c, "Notify", &tcinput.ToggleConf{Default: true})
		}, true, false)

	checkDefault(t, "color picker", "color_picker_component_Tint", "#00ff00",
		func(c *tgframe.Container) string {
			return tcinput.ColorPicker(c, "Tint",
				&tcinput.ColorPickerConf{Default: "#ff0000"})
		}, "#ff0000", "#00ff00")

	// The state holds the float64 a JSON number lands as, whatever T is.
	checkDefault(t, "number", "number_component_Age", 7.0,
		func(c *tgframe.Container) int64 {
			return tcinput.Number(c, "Age", &tcinput.NumberConf[int64]{Default: 30})
		}, 30, 7)

	checkDefault(t, "slider", "slider_component_Level", 70.0,
		func(c *tgframe.Container) int {
			return tcinput.Slider(c, "Level",
				(&tcinput.SliderConf[int]{}).SetDefault(40))
		}, 40, 70)

	checkDefault(t, "select slider", "select_slider_component_Size", 1,
		func(c *tgframe.Container) int {
			return tcinput.SelectSlider(c, "Size", fruits,
				&tcinput.SelectSliderConf{Default: 2})
		}, 2, 1)

	// Select's state is 1-based, so a stored 1 is items[0]; both the conf's
	// Default and the return stay 0-based.
	checkDefault(t, "select", "select_component_Fruit", 1,
		func(c *tgframe.Container) *int {
			return tcinput.Select(c, "Fruit", fruits,
				(&tcinput.SelectConf{}).SetDefault(1))
		}, ptr(1), ptr(0))

	checkDefault(t, "radio", "radio_component_Fruit", 0,
		func(c *tgframe.Container) *int {
			return tcinput.Radio(c, "Fruit", fruits,
				(&tcinput.RadioConf{}).SetDefault(1))
		}, ptr(1), ptr(0))

	checkDefault(t, "multiselect", "multiselect_component_Fruit", []int{1},
		func(c *tgframe.Container) []int {
			return tcinput.Multiselect(c, "Fruit", fruits,
				&tcinput.MultiselectConf{Default: []int{0, 2}})
		}, []int{0, 2}, []int{1})

	checkDefault(t, "datepicker", "datepicker_component_When", "2026-09-08",
		func(c *tgframe.Container) *time.Time {
			return tcinput.Datepicker(c, "When",
				(&tcinput.DatepickerConf{}).SetDefault(date(2026, 1, 2)))
		}, ptr(date(2026, 1, 2)), ptr(date(2026, 9, 8)))

	checkDefault(t, "timepicker", "datepicker_component_At", "09:30",
		func(c *tgframe.Container) *time.Time {
			return tcinput.Timepicker(c, "At",
				(&tcinput.TimepickerConf{}).SetDefault(clock(8, 15)))
		}, ptr(clock(8, 15)), ptr(clock(9, 30)))

	checkDefault(t, "datetimepicker", "datepicker_component_Now",
		"2026-09-08T09:30",
		func(c *tgframe.Container) *time.Time {
			return tcinput.Datetimepicker(c, "Now",
				(&tcinput.DatetimepickerConf{}).SetDefault(
					time.Date(2026, 1, 2, 8, 15, 0, 0, time.UTC)))
		}, ptr(time.Date(2026, 1, 2, 8, 15, 0, 0, time.UTC)),
		ptr(time.Date(2026, 9, 8, 9, 30, 0, 0, time.UTC)))
}

// TestClearingIsAnAnswer pins the other half of the Default rule, which is
// easy to state backwards: Default stands in only until the app user has
// answered, and emptying an input is an answer. Clearing does not put the
// default back — on any of these, which is what makes the rule one rule.
func TestClearingIsAnAnswer(t *testing.T) {
	// Mantine hands an emptied number box back as "", which reaches the state
	// as the 0 that Number() makes of it.
	t.Run("number", func(t *testing.T) {
		state := tgframe.NewState()
		state.Set("number_component_Count", 0.0)

		got := tcinput.Number(defaultContainer(state), "Count",
			&tcinput.NumberConf[int]{Default: 10})
		if got != 0 {
			t.Errorf("cleared Number = %v, want 0 rather than the default", got)
		}
	})

	t.Run("textbox", func(t *testing.T) {
		state := tgframe.NewState()
		state.Set("textbox_component_Name", "")

		got := tcinput.Textbox(defaultContainer(state), "Name",
			&tcinput.TextboxConf{Default: "toolgui"})
		if got != "" {
			t.Errorf("cleared Textbox = %q, want empty rather than the default", got)
		}
	})

	t.Run("datepicker", func(t *testing.T) {
		state := tgframe.NewState()
		state.Set("datepicker_component_When", "")

		got := tcinput.Datepicker(defaultContainer(state), "When",
			(&tcinput.DatepickerConf{}).SetDefault(date(2026, 1, 2)))
		if got != nil {
			t.Errorf("cleared Datepicker = %v, want nil rather than the default", *got)
		}
	})
}

// date is what a [tcinput.Datepicker] hands back for a day: midnight UTC.
func date(year int, month time.Month, day int) time.Time {
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}

// clock is what a [tcinput.Timepicker] hands back for a time of day: that
// clock on 1 January year 0 in UTC, the date time.Parse fills in for a layout
// that names none.
func clock(hour, min int) time.Time {
	return time.Date(0, time.January, 1, hour, min, 0, 0, time.UTC)
}

// TestPickersAgreeOnTheTimeType pins what unifying the three pickers on
// time.Time bought: a date and a time of day picked separately add up to the
// datetime a single picker would have given, with no conversion in between.
func TestPickersAgreeOnTheTimeType(t *testing.T) {
	state := tgframe.NewState()
	state.Set("datepicker_component_Day", "2026-09-08")
	state.Set("datepicker_component_At", "09:30")
	state.Set("datepicker_component_Both", "2026-09-08T09:30")

	c := defaultContainer(state)
	day := tcinput.Datepicker(c, "Day")
	at := tcinput.Timepicker(c, "At")
	both := tcinput.Datetimepicker(c, "Both")

	if day == nil || at == nil || both == nil {
		t.Fatalf("Datepicker = %v, Timepicker = %v, Datetimepicker = %v, "+
			"want all three set", day, at, both)
	}

	got := day.Add(time.Duration(at.Hour())*time.Hour +
		time.Duration(at.Minute())*time.Minute)
	if !got.Equal(*both) {
		t.Errorf("date + time of day = %v, want the datetime %v", got, *both)
	}
}

// TestPickersDropTheHalfTheyDoNotAsk pins that each picker keeps only its own
// half of a Default, so a default and a pick of the same moment compare equal
// rather than differing by whatever the caller's time.Time carried besides.
func TestPickersDropTheHalfTheyDoNotAsk(t *testing.T) {
	full := time.Date(2026, 9, 8, 9, 30, 45, 0, time.FixedZone("UTC+8", 8*3600))

	c := defaultContainer(tgframe.NewState())

	if got := tcinput.Datepicker(c, "When",
		(&tcinput.DatepickerConf{}).SetDefault(full)); !got.Equal(date(2026, 9, 8)) {
		t.Errorf("Datepicker default = %v, want the day at midnight UTC", *got)
	}

	if got := tcinput.Timepicker(c, "At",
		(&tcinput.TimepickerConf{}).SetDefault(full)); !got.Equal(clock(9, 30)) {
		t.Errorf("Timepicker default = %v, want the clock alone", *got)
	}

	// The datetime picker keeps both halves, and drops only the seconds its
	// wire format cannot carry.
	want := time.Date(2026, 9, 8, 9, 30, 0, 0, time.UTC)
	if got := tcinput.Datetimepicker(c, "Now",
		(&tcinput.DatetimepickerConf{}).SetDefault(full)); !got.Equal(want) {
		t.Errorf("Datetimepicker default = %v, want %v", *got, want)
	}
}

// TestSetAsInitialValue is the user path the state getters used to panic on:
// writing a component's key from Go code, with a Go type rather than the
// float64 an event would have landed, to give the input an initial value.
func TestSetAsInitialValue(t *testing.T) {
	t.Run("number", func(t *testing.T) {
		state := tgframe.NewState()
		state.Set("number_component_Age", 30)

		if got := tcinput.Number[int64](defaultContainer(state), "Age"); got != 30 {
			t.Fatalf("Number = %v, want 30", got)
		}
	})

	// Select stores a 1-based index, so 2 preselects the second item.
	t.Run("select", func(t *testing.T) {
		state := tgframe.NewState()
		state.Set("select_component_Fruit", 2)

		got := tcinput.Select(defaultContainer(state), "Fruit",
			[]string{"apple", "banana"})
		if got == nil || *got != 1 {
			t.Fatalf("Select = %v, want 1", got)
		}
	})

	// Radio stores a 0-based one, so the same item is 1.
	t.Run("radio", func(t *testing.T) {
		state := tgframe.NewState()
		state.Set("radio_component_Fruit", 1)

		got := tcinput.Radio(defaultContainer(state), "Fruit",
			[]string{"apple", "banana"})
		if got == nil || *got != 1 {
			t.Fatalf("Radio = %v, want 1", got)
		}
	})

	t.Run("textbox", func(t *testing.T) {
		state := tgframe.NewState()
		state.Set("textbox_component_Name", "toolgui")

		if got := tcinput.Textbox(defaultContainer(state), "Name"); got != "toolgui" {
			t.Fatalf("Textbox = %q, want toolgui", got)
		}
	})

	t.Run("checkbox", func(t *testing.T) {
		state := tgframe.NewState()
		state.Set("checkbox_component_Agree", true)

		if got := tcinput.Checkbox(defaultContainer(state), "Agree"); !got {
			t.Fatalf("Checkbox = %v, want true", got)
		}
	})
}

// TestWrongTypeDoesNotPanic pins the reverse: a key written with a type the
// component cannot use reads as unset, rather than taking the page down. What
// "unset" is depends on the component — the default for one that has one, nil
// for one that can genuinely have no value.
func TestWrongTypeDoesNotPanic(t *testing.T) {
	state := tgframe.NewState()
	state.Set("number_component_Age", "thirty")
	state.Set("select_component_Fruit", "banana")
	state.Set("textbox_component_Name", 1)

	c := defaultContainer(state)

	if got := tcinput.Number[int64](c, "Age"); got != 0 {
		t.Errorf("Number = %v, want 0", got)
	}
	if got := tcinput.Select(c, "Fruit", []string{"apple", "banana"}); got != nil {
		t.Errorf("Select = %v, want nil", *got)
	}
	if got := tcinput.Textbox(c, "Name"); got != "" {
		t.Errorf("Textbox = %q, want empty", got)
	}
}

// TestCheckboxDefault pins the three states a Default: true checkbox goes
// through. The middle one is the whole point: an explicit false has to beat
// the default, not read as "not touched yet".
//
// [TestToggleDefault] is the same set for the switch, which shares the
// behaviour but not the code.
func TestCheckboxDefault(t *testing.T) {
	conf := func() *tcinput.CheckboxConf {
		return &tcinput.CheckboxConf{Default: true}
	}

	t.Run("untouched", func(t *testing.T) {
		state := tgframe.NewState()

		if got := tcinput.Checkbox(defaultContainer(state), "Agree", conf()); !got {
			t.Fatalf("Checkbox = %v, want true", got)
		}
	})

	t.Run("unchecked", func(t *testing.T) {
		state := tgframe.NewState()
		state.Set("checkbox_component_Agree", false)

		if got := tcinput.Checkbox(defaultContainer(state), "Agree", conf()); got {
			t.Fatalf("Checkbox = %v, want false", got)
		}
	})

	t.Run("checked", func(t *testing.T) {
		state := tgframe.NewState()
		state.Set("checkbox_component_Agree", true)

		if got := tcinput.Checkbox(defaultContainer(state), "Agree", conf()); !got {
			t.Fatalf("Checkbox = %v, want true", got)
		}
	})

	// The default only stands in for a missing value, so a checkbox without
	// one still starts off.
	t.Run("no default", func(t *testing.T) {
		state := tgframe.NewState()

		if got := tcinput.Checkbox(defaultContainer(state), "Agree"); got {
			t.Fatalf("Checkbox = %v, want false", got)
		}
	})
}

func TestToggleDefault(t *testing.T) {
	conf := func() *tcinput.ToggleConf {
		return &tcinput.ToggleConf{Default: true}
	}

	t.Run("untouched", func(t *testing.T) {
		state := tgframe.NewState()

		if got := tcinput.Toggle(defaultContainer(state), "Notify", conf()); !got {
			t.Fatalf("Toggle = %v, want true", got)
		}
	})

	t.Run("switched off", func(t *testing.T) {
		state := tgframe.NewState()
		state.Set("toggle_component_Notify", false)

		if got := tcinput.Toggle(defaultContainer(state), "Notify", conf()); got {
			t.Fatalf("Toggle = %v, want false", got)
		}
	})

	t.Run("switched on", func(t *testing.T) {
		state := tgframe.NewState()
		state.Set("toggle_component_Notify", true)

		if got := tcinput.Toggle(defaultContainer(state), "Notify", conf()); !got {
			t.Fatalf("Toggle = %v, want true", got)
		}
	})

	t.Run("no default", func(t *testing.T) {
		state := tgframe.NewState()

		if got := tcinput.Toggle(defaultContainer(state), "Notify"); got {
			t.Fatalf("Toggle = %v, want false", got)
		}
	})
}
