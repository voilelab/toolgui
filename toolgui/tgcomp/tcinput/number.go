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
	Default     *T           `json:"default,omitempty"`
	Min         *T           `json:"min,omitempty"`
	Max         *T           `json:"max,omitempty"`
	Step        *T           `json:"step,omitempty"`
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

// NumberConf is the configuration for a number component.
type NumberConf[T Numeric] struct {
	// Default is the default value of the number component.
	Default *T

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

	// ID is the ID of the number component.
	ID string
}

func (c *NumberConf[T]) SetDefault(v T) *NumberConf[T] {
	c.Default = &v
	return c
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

// Number create a number input and return its value. conf may be nil.
func Number[T Numeric](c *tgframe.Container, label string, conf *NumberConf[T]) *T {
	if conf == nil {
		conf = &NumberConf[T]{}
	}

	comp := newNumberComponent[T](label)
	comp.Placeholder = conf.Placeholder
	comp.Color = conf.Color
	comp.Default = conf.Default
	comp.Min = conf.Min
	comp.Max = conf.Max
	comp.Step = conf.Step
	comp.Disabled = conf.Disabled

	if conf.ID != "" {
		comp.SetID(conf.ID)
	}

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
		return conf.Default
	}

	v := T(*val)
	return &v
}
