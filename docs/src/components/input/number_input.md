# Number Input

Number create a number input and return its value, or nil when the input
holds nothing the page can use.

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

## Min and Max are reported, then applied

The box does not enforce the range: it keeps whatever the app user typed,
marks itself invalid, shows a message beside itself and sends the value on as
it is. `Number` applies the range on arrival, so **the value a page gets is
always within `Min` and `Max`**.

That is the point of sending it. An out-of-range value used to be held back,
which left the server on the last one that happened to be inside the range —
so a button pressed while the box was red handed the page a number that was
no longer on screen, and pages set no `Min`/`Max` at all and clamped in Go by
hand instead. Now the value moves with what is typed: type 999 over a `Max` of
24 and the page reads 24, not whatever was there before.

The bounds are compared before an integral `T` truncates, on the number the
app user actually typed. A float no `T` can hold — a pasted `1e20` is no
`int` — has no number to report and no bound to be pulled to, so it reads as
`Default`.

## Example

```go
numberValue := tgcomp.Number(numberCompCol, "Number",
	(&tcinput.NumberConf[float64]{
		Placeholder: "input the value here",
		Color:       tcutil.ColorSuccess,
		Default:     10,
	}).SetMin(10).SetMax(20).SetStep(2))

// Out of range, numberValue is the bound, not the last value in range.
tgcomp.Text(numberCompCol, fmt.Sprint("Value: ", numberValue),
	&tgcomp.TextConf{ID: "number_result"})

if tgcomp.Button(numberCompCol, "Save number") && numberValue != nil {
	tgcomp.Text(numberCompCol, "Saved: "+valStr)
}
```
