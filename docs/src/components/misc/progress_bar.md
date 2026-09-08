# Progress Bar

ProgressBar is a component that displays a progress bar.

## API

### Interface

```go
func ProgressBar(c *tgframe.Container, value int, label string, conf ...*ProgressBarConf) *progressBarComponent
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

The returned component can be updated while the page function is running:

```go
func (p *progressBarComponent) SetValue(value int)
func (p *progressBarComponent) SetLabel(label string)
func (p *progressBarComponent) Remove()
```

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
