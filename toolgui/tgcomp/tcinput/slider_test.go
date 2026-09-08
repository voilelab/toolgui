package tcinput_test

import (
	"encoding/json"
	"testing"

	"github.com/voilelab/toolgui/toolgui/tgcomp/tcinput"
	"github.com/voilelab/toolgui/toolgui/tgframe"
)

// sliderProps reads back what the component put on the wire, which is the
// only place a defaulted min/max/step is visible.
func sliderProps(t *testing.T, pack tgframe.NotifyPack) struct {
	Min  float64 `json:"min"`
	Max  float64 `json:"max"`
	Step float64 `json:"step"`
} {
	t.Helper()

	bs, err := json.Marshal(pack)
	if err != nil {
		t.Fatal(err)
	}

	var got struct {
		Component struct {
			Min  float64 `json:"min"`
			Max  float64 `json:"max"`
			Step float64 `json:"step"`
		} `json:"component"`
	}
	if err := json.Unmarshal(bs, &got); err != nil {
		t.Fatal(err)
	}

	return got.Component
}

// TestSliderPanicsOnInvertedRange pins the ticket's rule that a parameter
// error is reported rather than silently corrected, the way the chart
// components already do it.
func TestSliderPanicsOnInvertedRange(t *testing.T) {
	for _, tc := range []struct {
		name string
		call func(*tgframe.Container)
	}{
		{"min above max", func(c *tgframe.Container) {
			tcinput.Slider(c, "s", (&tcinput.SliderConf[int]{}).SetMin(10).SetMax(0))
		}},
		// The defaulted end counts too: a min above the default max of 100 is
		// the same mistake written shorter.
		{"min above the default max", func(c *tgframe.Container) {
			tcinput.Slider(c, "s", (&tcinput.SliderConf[int]{}).SetMin(200))
		}},
		{"max below the default min", func(c *tgframe.Container) {
			tcinput.Slider(c, "s", (&tcinput.SliderConf[int]{}).SetMax(-1))
		}},
		{"negative step", func(c *tgframe.Container) {
			tcinput.Slider(c, "s", (&tcinput.SliderConf[int]{}).SetStep(-1))
		}},
		{"default outside the range", func(c *tgframe.Container) {
			tcinput.Slider(c, "s", (&tcinput.SliderConf[int]{}).
				SetMin(0).SetMax(10).SetDefault(11))
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Fatal("did not panic")
				}
			}()

			var packs []tgframe.NotifyPack
			tc.call(testContainer(tgframe.NewState(), &packs))
		})
	}

	// An equal min and max is degenerate but not a mistake: the slider has
	// one position and stays on it.
	t.Run("min equal to max", func(t *testing.T) {
		var packs []tgframe.NotifyPack
		got := tcinput.Slider(testContainer(tgframe.NewState(), &packs), "s",
			(&tcinput.SliderConf[int]{}).SetMin(5).SetMax(5))
		if got == nil || *got != 5 {
			t.Fatalf("got %v, want 5", got)
		}
	})
}

// TestSliderKeepsAStepThatDoesNotDivideTheRange records the other half of the
// ticket's parameter question: a step that does not fit a whole number of
// times into the range is not an error, so it is passed through as given and
// the ends of the range stay where the caller put them. What the last step
// does is the client's business.
func TestSliderKeepsAStepThatDoesNotDivideTheRange(t *testing.T) {
	var packs []tgframe.NotifyPack
	c := testContainer(tgframe.NewState(), &packs)

	tcinput.Slider(c, "s", (&tcinput.SliderConf[int]{}).
		SetMin(0).SetMax(10).SetStep(3))

	if len(packs) != 1 {
		t.Fatalf("got %d packs, want 1", len(packs))
	}

	got := sliderProps(t, packs[0])
	if got.Min != 0 || got.Max != 10 || got.Step != 3 {
		t.Errorf("min/max/step = %v/%v/%v, want 0/10/3",
			got.Min, got.Max, got.Step)
	}
}

// TestSliderDefaultsTheStep covers the step a conf that names none gets, and
// the zero step that reads as unset the way [tcinput.Number]'s does.
func TestSliderDefaultsTheStep(t *testing.T) {
	for _, tc := range []struct {
		name string
		call func(*tgframe.Container)
		want float64
	}{
		{"integral", func(c *tgframe.Container) {
			tcinput.Slider[int](c, "s")
		}, 1},
		// A hundredth of the default 0..100 range.
		{"float", func(c *tgframe.Container) {
			tcinput.Slider[float64](c, "s")
		}, 1},
		{"float over a narrow range", func(c *tgframe.Container) {
			tcinput.Slider(c, "s", (&tcinput.SliderConf[float64]{}).
				SetMin(0).SetMax(1))
		}, 0.01},
		{"explicit zero", func(c *tgframe.Container) {
			tcinput.Slider(c, "s", (&tcinput.SliderConf[int]{}).SetStep(0))
		}, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var packs []tgframe.NotifyPack
			tc.call(testContainer(tgframe.NewState(), &packs))

			if got := sliderProps(t, packs[0]).Step; got != tc.want {
				t.Errorf("step = %v, want %v", got, tc.want)
			}
		})
	}
}

// TestSliderDoesNotWriteBackToTheCallersConf is the same guarantee
// [tcinput.Number] carries: the defaults land on the component, so a conf
// reused across runs is left as the caller wrote it.
func TestSliderDoesNotWriteBackToTheCallersConf(t *testing.T) {
	conf := (&tcinput.SliderConf[int]{}).SetStep(0)

	var packs []tgframe.NotifyPack
	tcinput.Slider(testContainer(tgframe.NewState(), &packs), "s", conf)

	if *conf.Step != 0 {
		t.Errorf("conf.Step = %v, want 0", *conf.Step)
	}
	if conf.Min != nil || conf.Max != nil || conf.Default != nil {
		t.Errorf("conf gained min/max/default: %v/%v/%v",
			conf.Min, conf.Max, conf.Default)
	}
}

// TestSliderAlwaysHasAValue is what makes the returned pointer safe to
// dereference: the handle is always somewhere, so there is no unset position
// for a nil to stand for.
func TestSliderAlwaysHasAValue(t *testing.T) {
	t.Run("falls back to min", func(t *testing.T) {
		var packs []tgframe.NotifyPack
		got := tcinput.Slider(testContainer(tgframe.NewState(), &packs), "s",
			(&tcinput.SliderConf[int]{}).SetMin(10).SetMax(20))
		if got == nil || *got != 10 {
			t.Fatalf("got %v, want 10", got)
		}
	})

	t.Run("falls back to the default", func(t *testing.T) {
		var packs []tgframe.NotifyPack
		got := tcinput.Slider(testContainer(tgframe.NewState(), &packs), "s",
			(&tcinput.SliderConf[int]{}).SetMin(10).SetMax(20).SetDefault(15))
		if got == nil || *got != 15 {
			t.Fatalf("got %v, want 15", got)
		}
	})

	// The client sends a JSON number whatever T is, so an integral T
	// truncates it, as Number does.
	t.Run("reads the state", func(t *testing.T) {
		state := tgframe.NewState()
		state.Set("slider_component_s", 12.9)

		var packs []tgframe.NotifyPack
		got := tcinput.Slider(testContainer(state, &packs), "s",
			(&tcinput.SliderConf[int]{}).SetMin(10).SetMax(20))
		if got == nil || *got != 12 {
			t.Fatalf("got %v, want 12", got)
		}
	})
}
