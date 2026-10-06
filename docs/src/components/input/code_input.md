# Code Input

CodeInput create a code editor with syntax highlight and return its value.

It is built on [CodeMirror 6](https://codemirror.net/): line numbers, bracket
matching, undo history and search (`Ctrl/Cmd+F`) come with it.

## API

### Interface

```go
func CodeInput(c *tgframe.Container, label string, conf ...*CodeInputConf) string
```

### Parameters

* `c` is Parent container.
* `label` is the label for the code input.
* `conf` is an optional configuration, at most one.

```go
// CodeInputConf is the configuration for a code input.
type CodeInputConf struct {
	tgframe.Base // ID

	// Language is the language to highlight, leave empty to use `go`.
	// Supported: go, python, javascript, typescript, jsx, tsx, json, sql,
	// html, css, markdown, yaml, shell. "text" or any other value
	// leaves the code unhighlighted.
	Language string

	// Height is the number of lines shown. default value is 10.
	Height int

	// Default is the default value of the code input.
	Default string

	// ResetKey drops the app user's input and restores Default whenever it
	// changes, e.g. a hash of the file the code was filled from.
	ResetKey string
}
```

## Example

```go
{{#include ../../../demos/code_input.go:demo}}
```

<div data-toolgui-demo="code_input"></div>

## When the code is sent

Like [`Textarea`](textarea.md), typing does not rerun the page. The code is
sent when the editor loses focus, or at once on `Ctrl+Enter` (`Cmd+Enter` on
macOS). Leaving the editor without a change sends nothing.

## Keyboard

`Tab` indents rather than moving focus. To leave the editor by keyboard, press
`Esc` and then `Tab`.
