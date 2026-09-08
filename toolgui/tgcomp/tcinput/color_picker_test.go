package tcinput_test

import (
	"testing"

	"github.com/voilelab/toolgui/toolgui/tgcomp/tcinput"
	"github.com/voilelab/toolgui/toolgui/tgframe"
)

// TestColorPickerReturnsLowercaseHex covers the format the page is promised:
// "#rrggbb" in lowercase, whatever case the value arrived in, and the conf's
// default until something is picked.
func TestColorPickerReturnsLowercaseHex(t *testing.T) {
	t.Run("black until picked", func(t *testing.T) {
		got := tcinput.ColorPicker(defaultContainer(tgframe.NewState()), "c")
		if got != "#000000" {
			t.Errorf("got %q, want #000000", got)
		}
	})

	t.Run("the conf default until picked", func(t *testing.T) {
		got := tcinput.ColorPicker(defaultContainer(tgframe.NewState()), "c",
			&tcinput.ColorPickerConf{Default: "#FF3860"})
		if got != "#ff3860" {
			t.Errorf("got %q, want #ff3860", got)
		}
	})

	t.Run("the picked color", func(t *testing.T) {
		state := tgframe.NewState()
		state.Set("color_picker_component_c", "#23D160")

		if got := tcinput.ColorPicker(defaultContainer(state), "c"); got != "#23d160" {
			t.Errorf("got %q, want #23d160", got)
		}
	})

	// A key written by hand with something that is not a color reads as
	// unpicked, rather than reaching the page as a color it cannot use.
	t.Run("a state value that is not a color", func(t *testing.T) {
		state := tgframe.NewState()
		state.Set("color_picker_component_c", "red")

		got := tcinput.ColorPicker(defaultContainer(state), "c",
			&tcinput.ColorPickerConf{Default: "#ffffff"})
		if got != "#ffffff" {
			t.Errorf("got %q, want #ffffff", got)
		}
	})
}

// TestColorPickerPanicsOnAMalformedDefault is the parameter-error half: the
// caller's own literal is reported rather than corrected.
func TestColorPickerPanicsOnAMalformedDefault(t *testing.T) {
	for _, def := range []string{"red", "#fff", "ff3860", "#gggggg"} {
		t.Run(def, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Fatal("did not panic")
				}
			}()

			tcinput.ColorPicker(defaultContainer(tgframe.NewState()), "c",
				&tcinput.ColorPickerConf{Default: def})
		})
	}
}

// TestSelectSliderAlwaysLandsOnAnItem covers the index the page gets back:
// never nil, and never one the item list cannot answer.
func TestSelectSliderAlwaysLandsOnAnItem(t *testing.T) {
	items := []string{"S", "M", "L"}

	t.Run("the first item until moved", func(t *testing.T) {
		got := tcinput.SelectSlider(defaultContainer(tgframe.NewState()), "size", items)
		if got == nil || *got != 0 {
			t.Fatalf("got %v, want 0", got)
		}
	})

	t.Run("the conf default until moved", func(t *testing.T) {
		got := tcinput.SelectSlider(defaultContainer(tgframe.NewState()), "size",
			items, &tcinput.SelectSliderConf{Default: 2})
		if got == nil || *got != 2 {
			t.Fatalf("got %v, want 2", got)
		}
	})

	// 0-based on the wire, unlike Select, which keeps 0 for "nothing
	// selected".
	t.Run("the state index", func(t *testing.T) {
		state := tgframe.NewState()
		state.Set("select_slider_component_size", 1)

		got := tcinput.SelectSlider(defaultContainer(state), "size", items)
		if got == nil || *got != 1 {
			t.Fatalf("got %v, want 1", got)
		}
	})

	// An index left over from a run whose list was longer is dropped rather
	// than clamped, so the value never disagrees with where the handle is.
	t.Run("an index the list no longer has", func(t *testing.T) {
		state := tgframe.NewState()
		state.Set("select_slider_component_size", 7)

		got := tcinput.SelectSlider(defaultContainer(state), "size", items,
			&tcinput.SelectSliderConf{Default: 1})
		if got == nil || *got != 1 {
			t.Fatalf("got %v, want 1", got)
		}
	})
}

// TestSelectSliderPanicsOnAnUnslidableList pins the parameter errors: there is
// no position to put the handle on in either case.
func TestSelectSliderPanicsOnAnUnslidableList(t *testing.T) {
	for _, tc := range []struct {
		name  string
		items []string
		conf  *tcinput.SelectSliderConf
	}{
		{"no items", nil, nil},
		{"default past the end", []string{"S", "M"},
			&tcinput.SelectSliderConf{Default: 2}},
		{"negative default", []string{"S", "M"},
			&tcinput.SelectSliderConf{Default: -1}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Fatal("did not panic")
				}
			}()

			c := defaultContainer(tgframe.NewState())
			if tc.conf == nil {
				tcinput.SelectSlider(c, "size", tc.items)
				return
			}

			tcinput.SelectSlider(c, "size", tc.items, tc.conf)
		})
	}
}

// TestToggleMatchesCheckbox is the ticket's reverse condition, written as a
// test: the two differ in how they are drawn, not in what they return.
func TestToggleMatchesCheckbox(t *testing.T) {
	state := tgframe.NewState()
	state.Set("toggle_component_Agree", true)
	state.Set("checkbox_component_Agree", true)

	c := defaultContainer(state)
	if got, want := tcinput.Toggle(c, "Agree"), tcinput.Checkbox(c, "Agree"); got != want {
		t.Errorf("Toggle = %v, Checkbox = %v", got, want)
	}
}
