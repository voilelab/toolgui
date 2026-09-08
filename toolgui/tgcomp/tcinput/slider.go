package tcinput

import (
	"fmt"

	"github.com/voilelab/toolgui/toolgui/tgcomp/tcutil"
	"github.com/voilelab/toolgui/toolgui/tgframe"
)

var _ tgframe.Component = &sliderComponent[float64]{}
var _ tgframe.Component = &sliderComponent[int64]{}

var sliderComponentName = "slider_component"

// The range a slider spans when the conf names neither end.
const (
	defaultSliderMin = 0
	defaultSliderMax = 100
)

type sliderComponent[T Numeric] struct {
	*tgframe.BaseComponent

	Label    string `json:"label"`
	Default  T      `json:"default"`
	Min      T      `json:"min"`
	Max      T      `json:"max"`
	Step     T      `json:"step"`
	Disabled bool   `json:"disabled"`
}

func newSliderComponent[T Numeric](label string) *sliderComponent[T] {
	return &sliderComponent[T]{
		BaseComponent: &tgframe.BaseComponent{
			Name: sliderComponentName,
			ID:   tcutil.NormalID(sliderComponentName, label),
		},
		Label: label,
	}
}

// SliderConf is the configuration for a slider component.
type SliderConf[T Numeric] struct {
	tgframe.Base

	// Default is the value the slider starts at. Defaults to Min.
	Default *T

	// Min is the low end of the range. Defaults to 0.
	Min *T

	// Max is the high end of the range. Defaults to 100.
	Max *T

	// Step is the distance between two positions. Defaults to 1 for an
	// integral T, and to a hundredth of the range for a floating point one.
	Step *T

	// Disabled is the disabled state of the slider component.
	Disabled bool
}

func (c *SliderConf[T]) SetDefault(v T) *SliderConf[T] {
	c.Default = &v
	return c
}

func (c *SliderConf[T]) SetMin(v T) *SliderConf[T] {
	c.Min = &v
	return c
}

func (c *SliderConf[T]) SetMax(v T) *SliderConf[T] {
	c.Max = &v
	return c
}

func (c *SliderConf[T]) SetStep(v T) *SliderConf[T] {
	c.Step = &v
	return c
}

// orDefault reads an optional conf field, falling back to def when it is unset.
func orDefault[T Numeric](v *T, def T) T {
	if v == nil {
		return def
	}

	return *v
}

// Slider create a slider over a numeric range and return its value.
//
// A slider always sits somewhere in its range, so the return is never nil: it
// is the value the app user left it at, else Default, else Min.
//
// The value is reported when the handle is released rather than on every tick
// of a drag, so a drag across the range is one rerun rather than one per step.
//
// A range whose Min is above its Max, or a negative Step, is a mistake in the
// caller rather than a value to correct, and panics.
func Slider[T Numeric](
	c *tgframe.Container, label string, conf ...*SliderConf[T]) *T {

	cf := tgframe.OneConf("Slider", conf)

	comp := newSliderComponent[T](label)
	comp.Min = orDefault(cf.Min, T(defaultSliderMin))
	comp.Max = orDefault(cf.Max, T(defaultSliderMax))
	comp.Disabled = cf.Disabled
	tgframe.SetConfID(comp, cf)

	if comp.Min > comp.Max {
		panic(fmt.Sprintf("toolgui: Slider %q has min %v above max %v",
			label, comp.Min, comp.Max))
	}

	if cf.Step != nil && *cf.Step < T(0) {
		panic(fmt.Sprintf("toolgui: Slider %q has a negative step %v",
			label, *cf.Step))
	}

	// A zero step would leave the handle unable to move, so it reads as
	// unset, as it does on Number. The default is written into the component
	// rather than back into conf: the caller owns conf and may well reuse it.
	comp.Step = orDefault(cf.Step, T(0))
	if comp.Step == T(0) {
		comp.Step = defaultStep(comp.Min, comp.Max)
	}

	// The default is what the app user sees before touching anything, so it
	// has to be a position on the slider.
	comp.Default = orDefault(cf.Default, comp.Min)
	if comp.Default < comp.Min || comp.Default > comp.Max {
		panic(fmt.Sprintf("toolgui: Slider %q has default %v outside [%v, %v]",
			label, comp.Default, comp.Min, comp.Max))
	}

	c.AddComponent(comp)

	// The client sends every number back as a JSON number, so the state holds
	// a float64 whatever T is; T(*val) truncates it back for an integral T.
	val := c.State.GetFloat(comp.ID)
	if val == nil {
		return &comp.Default
	}

	v := T(*val)
	return &v
}

// defaultStep is the step a conf that names none gets: whole numbers for an
// integral T, and a hundredth of the range for a floating point one, which is
// about the resolution a drag across the track can pick out.
func defaultStep[T Numeric](min, max T) T {
	if isIntegral[T]() {
		return T(1)
	}

	step := (max - min) / T(100)
	if step == T(0) {
		return T(1)
	}

	return step
}
