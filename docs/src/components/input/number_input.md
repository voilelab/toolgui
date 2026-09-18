# Number Input

Number create a number input and return its value, always within `Min` and
`Max`, and whether that value is the one the app user entered.

## API

### Interface

```go
type Numeric interface {
	~int | ~int64 | ~float64
}

func Number[T Numeric](c *tgframe.Container, label string, conf ...*NumberConf[T]) (T, bool)
```

### Parameters

* `c` is Parent container.
* `label` is the label of the number input.
* `conf` is an optional configuration, at most one.

`Number` is one function for every numeric type it supports, so the type comes
from the conf or from an explicit instantiation:

```go
count, _ := tgcomp.Number[int](p.Main, "Count")
ratio, _ := tgcomp.Number(p.Main, "Ratio", &tcinput.NumberConf[float64]{})
```

Neither of those sets a bound, so the second return is always `true` there and
`_` is the honest way to write it. Read it wherever `Min` or `Max` is set —
see [The second return](#the-second-return) below.

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

An integral `T` still truncates a fractional value that is inside the range:
`20.9` under a `Max` of 21 reads as `20`. That is reported rather than
enforced too — see [The second return](#the-second-return).

## The second return

Pulling the value into the range keeps it usable, but it says nothing about
where it came from: a `Max` of 24 reads as 24 whether the app user typed 24 or
typed 999. Storing the second is the bug this input used to have, and clamping
did not fix it — it only made the number stored a fresh 24 instead of a stale
one. The app user still sees a number they never entered come back.

The second return is what tells those apart:

```go
limit, ok := tgcomp.Number(p.Main, "Limit", conf)
if !ok {
    tgcomp.Text(p.Main, "Enter a limit between 0 and 24.")
    return nil
}
save(limit)
```

It is `false` when what arrived was outside `Min` or `Max`, and when `T` does
not hold it as it is:

* a pasted `1e20` is no `int` — there is no number to report, so the value
  reads as `Default`;
* a typed `20.9` is no `int` either — an integral `T` truncates it to `20`,
  and nothing on the wire says `T` is integral, so the box takes a decimal
  whatever `T` is. The value is the `20`, and the signal says it is not what
  was typed.

Neither is the app user's number, any more than a clamped 999 is. With no
bounds set and a `T` that holds whatever arrives exactly, it is always `true`:
an untouched box reading as `Default` and an emptied one reading as zero are
both answers, not refusals.

**The value is still worth reading when it is `false`.** It is the nearest one
`T` holds inside the range, which is what a page that only wants to display
something should show. The signal is extra information, not a replacement for the value:
handing back the raw 999 instead would put back the implementation-defined
conversion an integral `T` does with a number it cannot hold.

A form with several bounded inputs `&&`s their signals together itself; there
is no form-level validity to read.

## Why a second return, and not something else

Three shapes were on the table. This is the one that is safe by default:
`x := Number(...)` stops compiling, so every call site has to look at the
signal once, rather than only the pages whose author already knew about the
problem. The cost is that `Number` is the only input with a comma-ok return —
an asymmetry, but a Go-idiomatic one, and one the bug is worth.

*A sibling query function* — `NumberInRange(state, id) bool`, in the shape of
`ButtonClicked` — was rejected for the opposite reason: it is opt-in. A page
that does not know it exists behaves exactly as it does today and stores the
clamped 24, so the default stays wrong. It would also have added another
component-layer function taking a `*tgframe.State`, which is the entry point
the component layer is trying to shed.

*A handle* — `Number` returning a value with `Value()` and `Valid()` on it —
carries the same signal and lines up with `StatusHandle`, but it is a heavy
shape for the simplest input there is, and it breaks existing call sites just
as hard as a second return does without buying anything more.

## Example

```go
{{#include ../../../demos/number_input.go:demo}}
```

<div data-toolgui-demo="number_input" data-toolgui-demo-height="640"></div>
