# Code

## API

### Interface

```go
func Code(c *tgframe.Container, code string, conf ...*CodeConf)
```

### Parameters

* `c`: Parent container.
* `code`: Code to display.
* `conf`: Optional configuration, at most one.

```go
// CodeConf provide extra config for Code Component.
type CodeConf struct {
	tgframe.Base // ID

	// Language is language of code block, leave empty to use `go`
	Language string
}
```

## Example

```go
tgcomp.Code(c, "package main\n\nfunc main() {\n\tprintln(\"Hello, World!\")")
```

```go
tgcomp.Code(c, "print('Hello, World!')",
    &tgcomp.CodeConf{
        Language: "python",
        ID:       "mycode",
    })
```
