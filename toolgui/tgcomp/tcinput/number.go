package tcinput

import (
	"github.com/voilelab/toolgui/toolgui/tgcomp/tcutil"
	"github.com/voilelab/toolgui/toolgui/tgframe"
)

var _ tgframe.Component = &numberComponent[float64]{}
var _ tgframe.Component = &numberComponent[int64]{}
var _ tgframe.Component = &numberComponent[int]{}

var numberComponentName = "number_component"

// Numeric is the value type a [Number] can hold. The tildes let a user's own
// named type be one, so a page can keep its domain type all the way in.
type Numeric interface {
	~int | ~int64 | ~float64
}

// isIntegral reports whether T counts in whole numbers. It is written as
// arithmetic rather than a type switch because a named type's dynamic type is
// itself, not the type it is defined from, so a switch would miss it.
func isIntegral[T Numeric]() bool {
	return T(1)/T(2) == T(0)
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

	// Default is what the input reads as before the app user types in it,
	// and what it reads as again once they empty it. The zero value is both
	// "no default" and a default of zero: the box starts empty either way,
	// and an empty box is zero, exactly as an empty [Textbox] is "".
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

// Number create a number input and return its value.
//
// There is no "nothing entered" state to report: an input nobody has typed in,
// and one that has been emptied, both read as Conf.Default.
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
	val := c.State.GetFloat(comp.ID)
	if val == nil {
		return cf.Default
	}

	return T(*val)
}
