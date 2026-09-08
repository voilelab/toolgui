# Status

Status reports a piece of work while the page function does it: a labelled
expander that collects the lines written to it, and closes as a success or a
failure.

## API

### Interface

```go
func Status(c *tgframe.Container, label string, conf ...*StatusConf) *StatusHandle
```

### Parameters

* `c`: Parent container.
* `label`: What the work is. An icon in front of it says how it is going.
* `conf`: Optional configuration, at most one.

```go
// StatusConf is the configuration for the Status component.
type StatusConf struct {
	tgframe.Base // ID

	// Expanded is whether the status starts open. It is closed by default:
	// the label says how the work is going, and the lines are the detail.
	Expanded bool
}
```

The returned handle is written while the page function runs:

```go
// Write appends a line to the status.
func (s *StatusHandle) Write(text string)

// Update replaces the label, leaving the state and the lines alone.
func (s *StatusHandle) Update(label string)

// Complete closes the status as a success, taking a new label at most one.
func (s *StatusHandle) Complete(label ...string)

// Fail closes the status as a failure, taking a new label at most one.
func (s *StatusHandle) Fail(label ...string)
```

`Complete` and `Fail` take at most one closing label; two or more panics, the
way passing two confs to a component does.

A status that is never closed stays in its running state, which is what the
page should show when the work did not get that far.

`StatusHandle` was called `StatusContainer`, and `Fail` was called `Error` —
a name that reads like the `error` interface. Both old names are kept as
deprecated aliases. See [what a component hands
back](../../architecture/components.md#what-a-component-hands-back).

## Example

```go
s := tgcomp.Status(c, "Importing…")

for _, f := range files {
	s.Write(f)
	if err := importFile(f); err != nil {
		s.Fail("Import failed: " + err.Error())
		return err
	}
}

s.Complete("Imported")
```

## How it is built

Status is an [Expand](../layout/expand.md) inside an
[Empty](../layout/empty.md) slot, redrawn on every change. That is why the
label may change without the expander closing: the expander keeps one id
through every redraw, which is the id `StatusConf.ID` sets, or the label the
status was created with.
