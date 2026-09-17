# Number Input

Number create a number input and return its value, or nil when the input
holds nothing the page can use.

## API

### Interface

```go
type Numeric interface {
	~int | ~int64 | ~float64
}

func Number[T Numeric](c *tgframe.Container, label string, conf ...*NumberConf[T]) *T
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
}

func (c *NumberConf[T]) SetDefault(v T) *NumberConf[T]
func (c *NumberConf[T]) SetMin(v T) *NumberConf[T]
func (c *NumberConf[T]) SetMax(v T) *NumberConf[T]
func (c *NumberConf[T]) SetStep(v T) *NumberConf[T]
```

An integral `T` cannot step by 0, so an explicit zero step means 1. The value
comes back from the client as a JSON number, so an integral `T` truncates it.

## Range and the nil value

`Min` and `Max` are reported, not enforced: the box keeps whatever the user
typed, marks itself invalid and sends the value on as it is. `Number` then
answers with nil rather than with the last value that was in range, so a
button pressed while the box is out of range cannot act on a number that is
no longer on screen.

`Number` returns nil in exactly these cases:

* the box is empty and the conf set no `Default`;
* what the user left in the box falls outside `Min` or `Max`.

An empty box is not a zero, and a value in range is unaffected -- a `Number`
with no `Min` and no `Max` never returns nil once something has been typed
into it. The range is judged on the value the page is handed, so an integral
`T` is truncated first: with `Max` 20, a typed 20.9 is the 20 that is in
range.

The `Default` is not a fallback for a refused value. It stands while the
input has sent nothing at all, which is what the box is showing; once the
user has made the box invalid, nothing is reported for it.

## Example

```go
numberValue := tgcomp.Number(numberCompCol, "Number",
	(&tcinput.NumberConf[float64]{
		Placeholder: "input the value here",
		Color:       tcutil.ColorSuccess,
	}).SetMin(10).SetMax(20).SetStep(2))

// Out of range, numberValue is nil rather than the last value in range.
valStr := "<none>"
if numberValue != nil {
	valStr = fmt.Sprint(*numberValue)
}

tgcomp.Text(numberCompCol, "Value: "+valStr,
	&tgcomp.TextConf{ID: "number_result"})

if tgcomp.Button(numberCompCol, "Save number") && numberValue != nil {
	tgcomp.Text(numberCompCol, "Saved: "+valStr)
}
```
