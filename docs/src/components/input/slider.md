# Slider

Slider create a slider over a numeric range and return its value.

## API

### Interface

```go
type Numeric interface {
	~int | ~int64 | ~float64
}

func Slider[T Numeric](c *tgframe.Container, label string, conf ...*SliderConf[T]) *T
```

### Parameters

* `c` is Parent container.
* `label` is the label of the slider.
* `conf` is an optional configuration, at most one.
* Return the value the slider sits at. Never nil.

`Slider` shares its constraint with [Number](number_input.md), so the type
comes from the conf or from an explicit instantiation:

```go
threshold := tgcomp.Slider[int](p.Main, "Threshold")
ratio := tgcomp.Slider(p.Main, "Ratio", &tcinput.SliderConf[float64]{})
```

A slider always sits somewhere in its range, so unlike `Number` it has no
"nothing entered" state and the returned pointer is never nil. It is the value
the app user left it at, else `Default`, else `Min`.

`SliderConf` is generic, so `tgcomp.SliderConf` is an alias you can name but
the methods below live on `tcinput.SliderConf`.
Import it from `github.com/voilelab/toolgui/toolgui/tgcomp/tcinput`.

```go
// SliderConf is the configuration for a slider component.
type SliderConf[T Numeric] struct {
	tgframe.Base // ID

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

func (c *SliderConf[T]) SetDefault(v T) *SliderConf[T]
func (c *SliderConf[T]) SetMin(v T) *SliderConf[T]
func (c *SliderConf[T]) SetMax(v T) *SliderConf[T]
func (c *SliderConf[T]) SetStep(v T) *SliderConf[T]
```

A zero `Step` reads as unset, as it does on `Number`: a slider that cannot move
is not a value to pass on. A `Step` that does not divide the range evenly is
fine, and leaves both ends of the range where you put them.

A `Min` above `Max`, a negative `Step`, or a `Default` outside the range are
mistakes in the call rather than values to correct, and panic.

## When the value is reported

The value is sent to Go when the handle is **released**, not while it is being
dragged. A drag across the range is one rerun rather than one per step it
crosses, which is what keeps a page with real work behind it usable.

The practical consequence is that anything downstream of a slider updates once
the app user lets go, not as they move. A value that has to follow the handle
live belongs in the browser, not in a rerun.

## Example

```go
threshold := tgcomp.Slider(p.Main, "Threshold",
	(&tcinput.SliderConf[int64]{}).SetMin(0).SetMax(100).SetStep(10).
		SetDefault(50))

// Never nil, so there is nothing to check first.
tgcomp.Text(p.Main, fmt.Sprint("Value: ", *threshold),
	&tgcomp.TextConf{ID: "slider_result"})
```

A slider derives its id from its label, so two with the same label collide.
Naming either of them is the way out:

```go
tgcomp.Slider[int](p.Main, "Weight")
tgcomp.Slider[int](p.Main, "Weight", &tcinput.SliderConf[int]{ID: "second_weight"})
```
