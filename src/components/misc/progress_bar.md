# Progress Bar

ProgressBar is a component that displays a progress bar.

## API

### Interface

```go
func ProgressBar(c *tgframe.Container, value int, label string) *progressBarComponent
```

### Parameters

* `c`: Parent container.
* `value`: Value of the progress bar, between 0 and 100.
* `label`: Label of the progress bar.

The returned component can be updated while the page function is running:

```go
func (p *progressBarComponent) SetValue(value int)
func (p *progressBarComponent) SetLabel(label string)
func (p *progressBarComponent) Remove()
```

## Example

```go
bar := tgcomp.ProgressBar(c, 50, "Progress")
for i := 0; i <= 100; i++ {
	bar.SetValue(i)
	time.Sleep(100 * time.Millisecond)
}

bar.SetLabel("Completed")
```
