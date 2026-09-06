# Code

## API

### Interface

```go
func Code(c *tgframe.Container, code string)
func CodeWithConf(c *tgframe.Container, code string, conf *CodeConf)
```

### Parameters

* `c`: Parent container.
* `code`: Code to display.
* `conf`: Configuration for the code component.

```go
// CodeConf provide extra config for Code Component.
type CodeConf struct {
	// Language is language of code block, leave empty to use `go`
	Language string

	// ID is id of the component
	ID string
}
```

## Example

```go
tgcomp.Code(c, "package main\n\nfunc main() {\n\tprintln(\"Hello, World!\")")
```

```go
tgcomp.CodeWithConf(c, "package main\n\nfunc main() {\n\tprintln(\"Hello, World!\")",
    &tgcomp.CodeConf{
        Language: "go",
        ID:       "mycode",
    })
```
