# Input Components

The input components provide UI for app-user to input their data.

```go
import "github.com/voilelab/toolgui/toolgui/tgcomp"
```

These components are shown on the `input` page of the demo app: [run it in
your browser](../../demo/#/input), or `task run_demo` and open
http://localhost:3000/input.

## `Default` and what you get back

Every input that takes a `Conf.Default` reads it the same way: it is what Go
gets back before the app user has touched the component, and it is what the
frontend renders on the first draw. One field, one meaning, everywhere.

The return type follows from one question — can this component have no value at
all? An input that always has one hands back a value; one that can genuinely be
empty hands back a pointer, and `nil` is that emptiness. There is no third case.

| Component | `Conf.Default` | Returns | Before any input |
| --- | --- | --- | --- |
| `Textbox` | `string` | `string` | `Default` |
| `Textarea` | `string` | `string` | `Default` |
| `Checkbox` | `bool` | `bool` | `Default` |
| `Toggle` | `bool` | `bool` | `Default` |
| `ColorPicker` | `string` | `string` | `Default`, else `#000000` |
| `Number` | `T` | `T` | `Default` |
| `Slider` | `*T` | `T` | `Default`, else `Min` |
| `SelectSlider` | `int` | `int` | `Default` |
| `MultiSelect` | `[]int` | `[]int` | `Default`, `nil` for none |
| `Select` | `*int` | `*int` | `Default`, `nil` for none |
| `Radio` | `*int` | `*int` | `Default`, `nil` for none |
| `DatePicker` | `*time.Time` | `*time.Time` | `Default`, `nil` for none |
| `TimePicker` | `*time.Time` | `*time.Time` | `Default`, `nil` for none |
| `DateTimePicker` | `*time.Time` | `*time.Time` | `Default`, `nil` for none |
| `FileUpload` | — | `*FileObject` | `nil` |

Two things follow from the table that are worth saying out loud:

* "Nothing selected" is `nil` for `Select`, `Radio` and `MultiSelect` alike, so
  the same check reads a single pick and a multiple one.
* The three pickers all hand back a `time.Time`, so a date and a time of day
  add up without a conversion in between. `DatePicker` keeps only the day, at
  midnight UTC; `TimePicker` keeps only the clock.

`FileUpload` is the one input with no `Default`: a browser refuses to have a
file input's value set from script, so a default would read back in Go while the
box on screen stayed empty.
