# Testing

`tgtest` runs a page inside a Go test, with no server and no browser. It
renders the page, sends the events a browser would, and lets the test read
back what the page drew.

```go
import "github.com/voilelab/toolgui/toolgui/tgtest"

func TestGreet(t *testing.T) {
    p := tgtest.Open(t, newApp(), "index")

    p.GetByLabel("Name").Input("Alice")
    p.GetByLabel("Greet").Click()

    if err := p.Err(); err != nil {
        t.Fatal(err)
    }

    if !p.HasText("Hello, Alice") {
        t.Error("no greeting")
    }
}
```

Each action waits for the run it starts, so what the test reads next is what
that run drew.

## Finding a component

| Method | Returns |
| --- | --- |
| `p.GetByLabel(label)` | The one component with that label; fails the test otherwise |
| `p.Get(id)` | The component of that id; fails the test without one |
| `p.FindByName(name)` | Every component of a type, e.g. `"form_component"` |
| `p.Find(match)` | Every component `match` accepts |
| `p.Main()`, `p.Sidebar()` | The root containers, to walk `Children` |

A `Node` carries the component's `Name`, `ID` and `Props` as sent to the
client. `p.HasText(s)` reports whether any text the page draws contains `s`;
props that never reach the screen as text, such as a fileupload's `accept` or
a link's `url`, don't count.

## Acting on it

| Method | Does |
| --- | --- |
| `n.Click()` | Click a button |
| `n.Input(v)` | Set a textbox, checkbox, number, ... to `v` |
| `n.Select(i)` | Pick item `i` (0-based) of a select, radio, select slider or menu; `-1` clears a select, and any other index outside the items fails the test |
| `n.SelectMany(i...)` | Pick items of a multiselect |
| `n.SelectKeys(k...)` | Pick rows of a DataFrame with row keys |
| `n.Upload(name, body)` | Pick a file in a fileupload |
| `n.UploadFiles(files...)` | Pick files in a multi-file upload |
| `n.Submit()` | Submit a form |
| `p.Rerun()` | Run the page again, like the rerun button |
| `p.Send(event)` | Send any `tgframe.Event` |

Actions on a disabled component fail the test, since a user cannot reach
them. Inside a form, inputs are held until the form is submitted, as in the
browser: by `Submit` on the form, or by clicking a button inside it.

## Errors

`p.Err()` is the error of the last run, `nil` when it succeeded. A failed run
leaves the components of the run before it in place, as the browser does.
