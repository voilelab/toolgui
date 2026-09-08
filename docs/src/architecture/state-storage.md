# State Storage

ToolGUI stores data at three levels, longest lived first:

* [App Cache](app-cache.md): for the whole process, shared by every user.
  Not provided by ToolGUI — the developer implements it.
* [Session Cache](session-cache.md): for one user's connection to the app.
* [State Cache](state-cache.md): for the page currently shown, through
  `p.State`.

Why cache at all:

- Faster access: Frequently used data can be retrieved from the cache much faster than recalculating it or fetching it from an external source every time. This improves the application's overall performance.

- Reduced resource usage: By avoiding redundant calculations and external data fetching, the app can conserve resources like CPU and network bandwidth.

## Reading a value out of the state

`p.State` holds `any`, so every getter has to answer what happens when the
value is not the type asked for. None of them panic: `GetString`, `GetFloat`
and `GetInt` return `nil`, `GetBool` returns `false`, and the generic
`State.Get[T]` returns the zero value and `false`.

```go
name, ok := p.State.Get[string]("name")
```

`State.Default[T]` is the same idea for a value the page mutates. It stores
the default the first time round and hands back a pointer, so what the page
writes through it is what the next run reads:

```go
todoList := p.State.Default("todoList", TODOList{})
todoList.Add("buy milk")
```

Numbers are the one place a type is not taken literally. The frontend sends
every number as JSON, so an event lands a `float64` whatever the component's
own type is, while a default written from Go code carries whichever integer
type was at hand. `GetFloat` and `GetInt` read either, so `Set(key, 30)`,
`Set(key, int64(30))` and `Set(key, 30.0)` are the same value. A string is
still not a number: `Set(key, "30")` reads back as `nil`.

## Setting an input's initial value

An input reads its value from the state under its own id, so writing that key
before the component runs is what the page reads back from it:

```go
func Main(p *tgframe.Params) error {
	if p.State.GetFloat("number_component_Age") == nil {
		p.State.Set("number_component_Age", 30)
	}

	age := tgcomp.Number[int64](p.Main, "Age") // 30, before anyone types
	...
}
```

The guard matters: `Set` on every run overwrites what the user just typed. Use
the getter that matches the stored value — `GetFloat` for a numeric key, since
`Get[float64]` would miss a default the page itself wrote as an `int`.

What `Set` does not do is fill in the field on screen. The state lives on the
server and is never sent to the client; the widget starts from the component's
own `Conf.Default`, and the browser only learns a value once someone enters
one. So a page that only writes the key reads 30 while showing an empty box.
To have both, write the key and set the conf:

```go
p.State.Set("number_component_Age", 30)
age := tgcomp.Number(p.Main, "Age", (&tgcomp.NumberConf[int64]{}).SetDefault(30))
```

Every component takes a conf, but only `Textbox`, `Number`, `Checkbox` and
`Multiselect` have a `Default` in theirs today, so `Select` and `Radio` can be
given a value the page reads but not one it shows.

The key is `<component name>_<label>`, unless the component was given an
explicit `ID` in its conf, in which case the key is that id verbatim.

| Component | Key | Stored value |
| --- | --- | --- |
| `Textbox` | `textbox_component_<label>` | `string` |
| `Textarea` | `textarea_component_<label>` | `string` |
| `Number` | `number_component_<label>` | any number |
| `Checkbox` | `checkbox_component_<label>` | `bool` |
| `Select` | `select_component_<label>` | item index, **1-based**; `0` is "nothing selected" |
| `Radio` | `radio_component_<label>` | item index, **0-based** |
| `Multiselect` | `multiselect_component_<label>` | item indices, **0-based**, as a list; `[]` is "nothing selected" |
| `Datepicker` | `datepicker_component_<label>` | `string`, `2006-01-02` |
| `Timepicker` | `datepicker_component_<label>` | `string`, `15:04` |
| `Datetimepicker` | `datepicker_component_<label>` | `string`, `2006-01-02T15:04` |

`Select` and `Radio` disagree on the base, which is the one asymmetry to watch
for. The frontend's select has a placeholder as its first option, so Go
numbers the real items from 1 and keeps 0 for "nothing selected"; radio has no
placeholder and numbers from 0, using an unset key for "nothing selected".
Neither shows through the API — both return a 0-based index, and `nil` when
nothing is selected — so it only bites when writing the key directly. To
preselect the second item:

```go
p.State.Set("select_component_Fruit", 2) // 1-based
p.State.Set("radio_component_Fruit", 1)  // 0-based
```

`Multiselect` numbers from 0 like `Radio`, and holds a list rather than one
index, so an empty list is a selection of nothing and an absent key is what
falls back to the conf's `Default`:

```go
p.State.Set("multiselect_component_Fruit", []int{0, 2})
```

The pickers parse the string they read, and a value in the wrong format fails
the run rather than being ignored, so write the format in the table exactly.
