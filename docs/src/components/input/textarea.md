# Textarea

Textarea create a textarea and return its value.

## API

### Interface

```go
func Textarea(c *tgframe.Container, label string, conf ...*TextareaConf) string
```

### Parameters

* `c` is Parent container.
* `label` is the label for textarea.
* `conf` is an optional configuration, at most one.

```go
// TextareaConf is the configuration for a textarea.
type TextareaConf struct {
	tgframe.Base // ID

	// Height is the height of the textarea. default value is 3.
	Height int

	// Default is the default value of the textarea.
	Default string

	// Color defines the color of the textarea
	Color tcutil.Color
}
```

## Example

```go
textareaValue := tgcomp.Textarea(p.Main, "Textarea",
	&tgcomp.TextareaConf{Height: 5})
tgcomp.Text(p.Main, "Value: "+textareaValue,
	&tgcomp.TextConf{ID: "textarea_result"})
```

![textarea component](textarea.png)
