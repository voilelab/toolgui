package tcinput_test

import (
	"testing"

	"github.com/voilelab/toolgui/toolgui/tgcomp/tcinput"
	"github.com/voilelab/toolgui/toolgui/tgframe"
)

func defaultContainer(state *tgframe.State) *tgframe.Container {
	return tgframe.NewContainer("test", state, func(tgframe.NotifyPack) {})
}

// TestSetAsInitialValue is the user path the state getters used to panic on:
// writing a component's key from Go code, with a Go type rather than the
// float64 an event would have landed, to give the input an initial value.
func TestSetAsInitialValue(t *testing.T) {
	t.Run("number", func(t *testing.T) {
		state := tgframe.NewState()
		state.Set("number_component_Age", 30)

		got := tcinput.Number[int64](defaultContainer(state), "Age")
		if got == nil || *got != 30 {
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
// component cannot use reads as unset, rather than taking the page down.
func TestWrongTypeDoesNotPanic(t *testing.T) {
	state := tgframe.NewState()
	state.Set("number_component_Age", "thirty")
	state.Set("select_component_Fruit", "banana")
	state.Set("textbox_component_Name", 1)

	c := defaultContainer(state)

	if got := tcinput.Number[int64](c, "Age"); got != nil {
		t.Errorf("Number = %v, want nil", *got)
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
