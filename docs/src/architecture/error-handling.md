# Error Handling

A page function has three ways a failure reaches the user:

```go
func Page(p *tgframe.Params) error {
    // 1. return it
    if err := doSomething(); err != nil {
        return err
    }

    // 2. a component records it and carries on
    tgcomp.Table(p.Main, head, table) // rows that don't match the head

    // 3. panic
    tgcomp.Column(p.Main, 0) // no such thing as zero columns

    return nil
}
```

All three end up in the same red message box under the page, and none of them
takes the server down. What differs is how much of the page survives.

## Which one to use

**`return err`** is for a failure of the job the page is doing: a file that
does not parse, a request that timed out, a form value the app rejects. The
page function decides when the run is over and hands the reason back. Nothing
after the return renders.

**`Container.Fail`** is for a failure the run's data decides — the same call
would be right on other data. A table whose rows do not match its head, an
image that will not encode, a date in the state that no longer parses. The run
records the error and keeps going: the component leaves a visible error
placeholder where it would have been, everything after it still renders, and
`App.Run` returns the failure once the page function is done.

```go
// Table create a table by heading(head) and values(table).
func Table(c *tgframe.Container, head []string, table [][]string, conf ...*TableConf) {
    // ...
    if len(table[0]) != len(head) {
        c.Fail(tgutil.NewError("len of head should equal to len of table[0]"))
        return
    }
    // ...
}
```

`Fail` is public, so a third-party component reports a bad value the same way
the built-in ones do. A nil error is nothing to report and does nothing.

**`panic(err)`** is for a misuse of the API: a call that no data could make
right. `Column(c, 0)` asks for zero columns, `OneConf` was handed two confs,
`Echo` cannot see its caller. These are programming errors the caller cannot
recover from, so the components panic instead of returning an error and keep
their signature small:

```go
// Column create N columns.
func Column(c *tgframe.Container, n uint, conf ...*ColumnConf) []*tgframe.Container {
    if n == 0 {
        panic("number of columns should > 0")
    }
    ...
}
```

The same rule applies to `App.AddPage` and `App.AddPageByConfig`: they panic
on a bad page config, because an app that cannot register its own pages has
nothing to run.

Application code is free to panic too — it just gets reported as a run error
rather than crashing the process.

## Where each component sits

The split is between what the call says and what the data says. A wrong
argument type or an out-of-range constant is the call; a length that only this
run's rows have is the data.

| Reports through `Fail` | Why |
| --- | --- |
| `Table`, head and row lengths | the rows are data |
| `DataFrame`, head, row, column conf and page size checks | same |
| `Chart` and friends, series against labels and points against kind | same |
| `JSON`, a value that will not marshal or a string that is not JSON | same |
| `Image`, a PNG or JPEG that will not encode | same |
| `Datepicker`, `Timepicker`, `DatetimePicker`, a stored value that will not parse | the state is data, and the browser or an old session may have written it |
| `Fileupload`, a stored pick that will not unmarshal | same |

| Still panics | Why |
| --- | --- |
| `Column(c, 0)`, `EqColumn` with an unsupported count | no data makes zero columns valid |
| `OneConf` handed two confs | the call passed two, not the data |
| `Echo` that cannot find its caller | the code is not shaped the way `Echo` needs |
| `Slider`, `SelectSlider`, `ColorPicker` bad bounds and defaults | the conf is the call |
| `Status` with more than one closing label | the call again |
| `Image` with an unsupported format constant or an unsupported argument type | the call again |
| `ChartKind`, `ColumnType`, `ColumnAlign` `String()` on an unknown constant | the call again |
| `App.AddPage` on a bad page config | an app that cannot register a page has nothing to run |

`session.go` panics with `ErrUpdateInterrupt` for neither reason: it is
control flow, the only way to unwind a run the next event has already made
stale, from wherever in the page function it happens to be. See
[Sentinel errors](#sentinel-errors) below.

## What happens to a returned error

`App.Run` calls the page function and wraps whatever comes back:

```go
err := pageFunc(&Params{...})
if err != nil {
    return tgutil.Errorf("%w", err)
}
```

A page function that returns nothing still fails the run when a component did:
`App.Run` returns what `Fail` recorded. The first failure of a run is the one
kept — a later one does not overwrite it — but every one of them leaves its own
placeholder, so the page shows all of them even though the caller reads one.
An error the page function returns itself wins over both: the page said the run
was over, and that reason is the one that comes back.

`tgutil.Errorf` and `tgutil.NewError` prefix the message with the name of the
function that created the error, so the log line says where it came from.
They wrap with `%w`, so `errors.Is` still works on the original error.

The `Session` turns the error into a result pack:

```go
err := s.app.RunWithHandlingPanic(s.pageName, s.state, sendNotifyPack)
if err != nil {
    s.sendResult(&ResultPack{Error: err.Error()})
    slog.Error("run err", "error", err)
    return
}

s.sendResult(&ResultPack{Success: true})
```

The client renders `ResultPack.Error` in `AppError`, a `is-danger` message
below the page body. Components the run already created stay on screen: the
error is appended to a half-drawn page, not a replacement for it. The next
run clears it.

## What happens to a panic

`RunWithHandlingPanic` recovers it and turns it into an error wrapping
`ErrPanic`:

```go
defer func() {
    r := recover()
    if r != nil {
        log.Println("Panic", r)
        err = tgutil.Errorf("%w: %v", ErrPanic, r)
    }
}()
```

From there it follows the path above, so the user sees the same red box, its
message being `panic: ` followed by the recovered value. A panic in one run does not affect the
session, the state, or the other users of a web executor.

Note that only panics inside the page function are covered. A panic in a
goroutine the page function started has no recover on its stack and kills the
process, as it does in any Go program.

## Sentinel errors

| Error | Meaning |
| --- | --- |
| `tgframe.ErrPageNotFound` | `App.Run` or `NewSession` got a name no page is registered under. |
| `tgframe.ErrPanic` | The page function panicked. Wraps the recovered value. |
| `tgframe.ErrUpdateInterrupt` | The run was cut short by a new event. Not an application error. |
| `tgframe.ErrDuplicatedID` | Two components of one run claimed the same id. Recorded through the same slot as `Fail`, so the first of the two comes back. |

`ErrUpdateInterrupt` is how an interrupted run unwinds. When an event arrives
while a page function is still running, the session sets a stop flag, and the
next component the page func creates panics with `ErrUpdateInterrupt` instead
of sending its notify pack:

```go
sendNotifyPack := func(pack NotifyPack) {
    if s.stopUpdating.Load() {
        panic(ErrUpdateInterrupt)
    }
    ...
}
```

`RunWithHandlingPanic` recovers it like any other panic, so the interrupted
run does log a `run err` line and does send an error result pack. The client
then receives the ready pack of the new run, which clears the error before it
is ever painted. Treat those log lines as noise from a rerun, not as failures.

## Outside the page function

The rest of the framework does not panic on the request path; it returns
errors and lets the executor decide.

* `Session.HandleRawEvent` reports a malformed event to the client and returns
  the error, so the executor can log it. A closed session ignores events
  instead of erroring.
* `Session` never fails a run because the client is gone. A send that fails is
  logged and dropped — there is nowhere left to report it to.
* The web executor answers with an HTTP status on the upload and page handlers,
  and sends a `ResultPack` over the socket when a session cannot be created.
* The desktop (Wails) backend returns the error to the frontend from its bound
  methods, `ErrNoSession` among them.

Package-level initialization is the one place the library panics on something
that is not a caller mistake: `toolguiweb.GetRootAssets` panics if the embedded
frontend assets cannot be read, since a build without them cannot serve
anything.
