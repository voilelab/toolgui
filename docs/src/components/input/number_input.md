# Number Input

Number create a number input and return its value.

## API

### Interface

```go
type Numeric interface {
	~int | ~int64 | ~float64
}

func Number[T Numeric](c *tgframe.Container, label string, conf ...*NumberConf[T]) T
```

### Parameters

* `c` is Parent container.
* `label` is the label of the number input.
* `conf` is an optional configuration, at most one.

`Number` is one function for every numeric type it supports, so the type comes
from the conf or from an explicit instantiation:

```go
count := tgcomp.Number[int](p.Main, "Count")
ratio := tgcomp.Number(p.Main, "Ratio", &tcinput.NumberConf[float64]{})
```

The `~` in the constraint lets a page keep its own named type all the way in,
so `type Rating int` is a `Number[Rating]`.

`NumberConf` is generic, so `tgcomp.NumberConf` is an alias you can name but
the methods below live on `tcinput.NumberConf`.
Import it from `github.com/voilelab/toolgui/toolgui/tgcomp/tcinput`.

```go
// NumberConf is the configuration for a number component.
type NumberConf[T Numeric] struct {
	tgframe.Base // ID

	// Default is what the input reads as before the app user types in it,
	// and what it reads as again once they empty it.
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

func (c *NumberConf[T]) SetMin(v T) *NumberConf[T]
func (c *NumberConf[T]) SetMax(v T) *NumberConf[T]
func (c *NumberConf[T]) SetStep(v T) *NumberConf[T]
```

An integral `T` cannot step by 0, so an explicit zero step means 1. The value
comes back from the client as a JSON number, so an integral `T` truncates it.

There is no "nothing entered" state to report: an input nobody has typed in
reads as `Default`. A zero `Default` is also "no default" — the box starts empty
either way, and an empty box is zero, exactly as an empty
[Textbox](textbox.md) is `""`. `Min`, `Max` and `Step` stay pointers, because
there a zero is a bound and an absent one is not.

Emptying the box afterwards is an answer of zero, not a return to `Default`:
the same rule [Textbox](textbox.md) and the pickers follow, where clearing
reads as `""` and as nil rather than putting the default back.

## Example

```go
numberValue := tgcomp.Number(numberCompCol, "Number",
	(&tcinput.NumberConf[float64]{
		Placeholder: "input the value here",
		Color:       tcutil.ColorSuccess,
		Default:     10,
	}).SetMin(10).SetMax(20).SetStep(2))

tgcomp.Text(numberCompCol, fmt.Sprint("Value: ", numberValue),
	&tgcomp.TextConf{ID: "number_result"})
```
