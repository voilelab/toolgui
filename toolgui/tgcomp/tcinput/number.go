package tcinput

import (
	"math"

	"github.com/voilelab/toolgui/toolgui/tgcomp/tcutil"
	"github.com/voilelab/toolgui/toolgui/tgframe"
)

var _ tgframe.Component = &numberComponent[float64]{}
var _ tgframe.Component = &numberComponent[int64]{}
var _ tgframe.Component = &numberComponent[int]{}

var numberComponentName = "number_component"

// Numeric is the value type a [Number] can hold. It is the state's own
// [tgframe.Numeric]: what a number input holds is what the state reads back,
// so there is one set of types, not two that have to be kept in step.
type Numeric = tgframe.Numeric

// isIntegral reports whether T counts in whole numbers. It is written as
// arithmetic rather than a type switch because a named type's dynamic type is
// itself, not the type it is defined from, so a switch would miss it.
func isIntegral[T Numeric]() bool {
	return T(1)/T(2) == T(0)
}

// holds reports whether T can hold f. A float outside an integral T's range
// converts to an implementation-defined value -- on amd64 a submitted 1e20
// lands on math.MinInt -- so a conversion is not enough on its own. The round
// trip catches it whatever that value is: only a float T really holds
// converts back to the one it was truncated from.
func holds[T Numeric](f float64) bool {
	return !math.IsNaN(f) && (!isIntegral[T]() || float64(T(f)) == math.Trunc(f))
}

type numberComponent[T Numeric] struct {
	*tgframe.BaseComponent

	Label       string       `json:"label"`
	Default     T            `json:"default,omitzero"`
	Min         *T           `json:"min,omitzero"`
	Max         *T           `json:"max,omitzero"`
	Step        *T           `json:"step,omitzero"`
	Color       tcutil.Color `json:"color"`
	Placeholder string       `json:"placeholder"`
	Disabled    bool         `json:"disabled"`
}

func newNumberComponent[T Numeric](label string) *numberComponent[T] {
	return &numberComponent[T]{
		BaseComponent: &tgframe.BaseComponent{
			Name: numberComponentName,
			ID:   tcutil.NormalID(numberComponentName, label),
		},
		Label: label,
	}
}

// NumberConf is the configuration for a number component. A generic conf
// embeds Base like any other.
type NumberConf[T Numeric] struct {
	tgframe.Base

	// Default is what the input reads as before the app user has typed in it.
	// The zero value is both "no default" and a default of zero: the box
	// starts empty either way, and an empty box is zero, exactly as an empty
	// [Textbox] is "".
	//
	// Emptying the box afterwards is an answer of zero, not a return to
	// Default — the same way clearing a [Textbox] reads as "" and clearing a
	// [DatePicker] reads as nil.
	Default T

	// Min is the minimum value of the number component.
	Min *T

	// Max is the maximum value of the number component.
	Max *T

	// Step is the step of the number component.
	Step *T

	// Color is the color of the number component.
	Color tcutil.Color

	// Placeholder is the placeholder of the number component.
	Placeholder string

	// Disabled is the disabled state of the number component.
	Disabled bool
}

func (c *NumberConf[T]) SetMin(v T) *NumberConf[T] {
	c.Min = &v
	return c
}

func (c *NumberConf[T]) SetMax(v T) *NumberConf[T] {
	c.Max = &v
	return c
}

func (c *NumberConf[T]) SetStep(v T) *NumberConf[T] {
	c.Step = &v
	return c
}

// Number create a number input and return its value, which is always within
// Conf.Min and Conf.Max.
//
// The input reports a value outside that range rather than enforcing it, so
// the app user keeps seeing what they typed, with the message beside it. What
// they typed is what comes back, so the value here is pulled into the range
// on arrival -- never the last one that happened to be inside it, which the
// page would read as what is on screen now.
//
// There is no "nothing entered" state to report: an input nobody has typed in
// reads as Conf.Default, and one the app user has emptied reads as zero.
func Number[T Numeric](c *tgframe.Container, label string, conf ...*NumberConf[T]) T {
	cf := tgframe.OneConf("Number", conf)

	comp := newNumberComponent[T](label)
	comp.Placeholder = cf.Placeholder
	comp.Color = cf.Color
	comp.Default = cf.Default
	comp.Min = cf.Min
	comp.Max = cf.Max
	comp.Step = cf.Step
	comp.Disabled = cf.Disabled
	tgframe.SetConfID(comp, cf)

	// An integral input cannot step by 0, so an explicit zero step means 1.
	// Written into the component rather than back into conf: the caller owns
	// conf and may well reuse it across runs.
	if isIntegral[T]() && comp.Step != nil && *comp.Step == T(0) {
		v := T(1)
		comp.Step = &v
	}

	c.AddComponent(comp)

	// The client sends every number back as a JSON number, so the state holds
	// a float64 whatever T is; T(*val) truncates it back for an integral T.
	val, ok := c.State.GetNumber[float64](comp.ID)
	if !ok {
		return cf.Default
	}

	// The bounds are compared in float64, before the truncation: that is the
	// number the app user typed, and it is the only form an out-of-range one
	// survives in -- converting first would land on whatever an integral T
	// does with a value it cannot hold.
	if comp.Min != nil && val < float64(*comp.Min) {
		return *comp.Min
	}

	if comp.Max != nil && val > float64(*comp.Max) {
		return *comp.Max
	}

	// In range, or unbounded. A float no T can hold is left to the Default:
	// there is no number to report and, with no bound to pull it to, nothing
	// to pull it to either.
	if !holds[T](val) {
		return cf.Default
	}

	return T(val)
}
