# Form

Form create a form component.
The input components in this component will not auto trigger script execution.
The user need to click the submit button to trigger the script execution.

## API

### Interface

```go
func Form(c *tgframe.Container, conf ...*FormConf) *tgframe.Container
```

### Parameters

* `c` is Parent container.
* `conf` is an optional configuration, at most one.

```go
// FormConf is the configuration for the Form component.
type FormConf struct {
	tgframe.Base // ID
}
```

## Example

```go
var a, b *float64
tgcomp.Form(formCompCol).With(func(c *tgframe.Container) {
	a = tgcomp.Number[float64](c, "a")
	b = tgcomp.Number[float64](c, "b")
})

if a != nil && b != nil {
	tgcomp.Text(formCompCol, fmt.Sprintf("int(a) + int(b) = %d", int(*a)+int(*b)))
}
```
