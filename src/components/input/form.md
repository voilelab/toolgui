# Form

Form create a form component.
The input components in this component will not auto trigger script execution.
The user need to click the submit button to trigger the script execution.

## API

### Interface

```go
func Form(c *tgframe.Container, id string) *tgframe.Container
```

### Parameters

* `c` is Parent container.
* `id` is the ID of the form.

## Example

```go
var a, b *float64
tgcomp.Form(formCompCol, "form").With(func(c *tgframe.Container) {
	a = tgcomp.NumberFloat64(p.State, c, "a")
	b = tgcomp.NumberFloat64(p.State, c, "b")
})

if a != nil && b != nil {
	tgcomp.Text(formCompCol, fmt.Sprintf("int(a) + int(b) = %d", int(*a)+int(*b)))
}
```
