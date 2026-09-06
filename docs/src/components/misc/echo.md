# Echo

Echo is a component that execute the code of lambda function and display the code of lambda function.

## API

```go
func Echo(c *tgframe.Container, code string, lambda func())
```

### Parameters

* `c`: Parent container.
* `code`: Code of current file. The user should use go embed syntax to embed the code.
* `lambda`: Lambda function to execute.

## Example

```go
//go:embed main.go
var code string

tgcomp.Echo(c, code, func() {
	println("Hello, World!")
})
```
