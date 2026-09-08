# Progress Bar

ProgressBar is a component that displays a progress bar.

## API

### Interface

```go
func ProgressBar(c *tgframe.Container, value int, label string, conf ...*ProgressBarConf) *ProgressBarHandle
```

### Parameters

* `c`: Parent container.
* `value`: Value of the progress bar, between 0 and 100.
* `label`: Label of the progress bar.
* `conf`: Optional configuration, at most one.

```go
// ProgressBarConf is the configuration for the ProgressBar component.
type ProgressBarConf struct {
	tgframe.Base // ID
}
```

The returned handle moves the bar along while the page function runs:

```go
func (p *ProgressBarHandle) SetValue(value int)
func (p *ProgressBarHandle) SetLabel(label string)
func (p *ProgressBarHandle) Remove()
```

`ProgressBarHandle` is exported, so the work that moves the bar need not be
the code that drew it: the handle can be kept in a struct field, passed to a
function, or named in an interface of your own. See [what a component hands
back](../../architecture/components.md#what-a-component-hands-back).

`Remove` also gives the bar's `Conf.ID` back, so the same id can be declared
again later in the run and the removed bar's state is not handed to whatever
lands on that id next.

## Example

```go
bar := tgcomp.ProgressBar(c, 50, "Progress")
for i := 0; i <= 100; i++ {
	bar.SetValue(i)
	time.Sleep(100 * time.Millisecond)
}

bar.SetLabel("Completed")
```
