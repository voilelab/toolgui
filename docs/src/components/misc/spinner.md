# Spinner

Spinner shows that the page function is busy, and returns the function that
takes it down again.

## API

### Interface

```go
func Spinner(c *tgframe.Container, label string, conf ...*SpinnerConf) func()
```

### Parameters

* `c`: Parent container.
* `label`: Text shown beside the spinner.
* `conf`: Optional configuration, at most one.

```go
// SpinnerConf is the configuration for the Spinner component.
type SpinnerConf struct {
	tgframe.Base // ID
}
```

The spinner sits in an [Empty](../layout/empty.md) slot, so taking it down
leaves the page as if it had never been there. The returned function may be
called more than once; only the first call does anything.

Taking it down is the only thing a spinner is asked to do afterwards, so it
hands back a bare `func()` rather than a handle. See [what a component hands
back](../../architecture/components.md#what-a-component-hands-back).

## Example

```go
stop := tgcomp.Spinner(c, "Working…")
rows := query()
stop()

tgcomp.Table(c, head, rows)
```

Call it with `defer` and the spinner is taken down however the work ends, a
panic included:

```go
defer tgcomp.Spinner(c, "Loading…")()
```
