package tcinput

import (
	"github.com/voilelab/toolgui/toolgui/tgcomp/tcutil"
	"github.com/voilelab/toolgui/toolgui/tgframe"
)

var _ tgframe.Component = &numberComponent[float64]{}
var _ tgframe.Component = &numberComponent[int64]{}

var numberComponentName = "number_component"

type numberComponent[T float64 | int64] struct {
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

func newNumberComponent[T float64 | int64](label string) *numberComponent[T] {
	return &numberComponent[T]{
		BaseComponent: &tgframe.BaseComponent{
			Name: numberComponentName,
			ID:   tcutil.NormalID(numberComponentName, label),
		},
		Label: label,
	}
}

// NumberConf is the configuration for a number component.
type NumberConf[T float64 | int64] struct {
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

func Number[T float64 | int64](s *tgframe.State, c *tgframe.Container, label string) *T {
	return NumberWithConf[T](s, c, label, nil)
}

func NumberWithConf[T float64 | int64](s *tgframe.State, c *tgframe.Container, label string, conf *NumberConf[T]) *T {
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

	// special case for int64
	switch any(T(0)).(type) {
	case int64:
		if conf.Step != nil && *conf.Step == T(0) {
			v := T(1)
			conf.Step = &v
		}
	}

	c.AddComponent(comp)

	val := s.GetFloat(comp.ID)
	if val == nil {
		return conf.Default
	}

	v := T(*val)
	return &v
}
