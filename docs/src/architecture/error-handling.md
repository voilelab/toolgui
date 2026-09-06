# Error Handling

A page function has two ways to report a failure:

```go
func Page(p *tgframe.Params) error {
    // 1. return it
    if err := doSomething(); err != nil {
        return err
    }

    // 2. panic
    tgcomp.Table(p.Main, head, table) // panics if len(head) != len(table[0])

    return nil
}
```

Both end up in the same place: the run stops, and the client shows a red
message box under the page. Neither of them takes the server down.

## Which one to use

**`return err`** is for a failure of the job the page is doing: a file that
does not parse, a request that timed out, a form value the app rejects.
The page function decides when the run is over and hands the reason back.

**`panic(err)`** is for a misuse of the API: an argument a component cannot
work with at all. Every panic in `tgcomp` is of this kind — a wrong column
count, a mismatched table header, an unsupported image type. Since these are
programming errors that a caller cannot recover from anyway, the components
panic instead of returning an error, and keep their signature small:

```go
// Column create N columns.
func Column(c *tgframe.Container, id string, n uint) []*tgframe.Container {
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

## What happens to a returned error

`App.Run` calls the page function and wraps whatever comes back:

```go
err := pageFunc(&Params{...})
if err != nil {
    return tgutil.Errorf("%w", err)
}
```

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
