# Form

Form create a form component.
The input components in this component will not auto trigger script execution.
The user need to submit the form to trigger the script execution.

A form is submitted by its built-in submit button, or by a `Button` written
inside it: a click inside a form is sent together with the values held since
the last submit, so the run that sees the click is the one that reads the new
values. That is what `HideSubmit` is for — a search form can offer its own
"Search" button instead of one of those beside a hardwired "Submit".

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

	// SubmitLabel is the text on the built-in submit button. Empty is
	// "Submit".
	SubmitLabel string

	// HideSubmit drops the built-in submit button. Submitting is then up to a
	// Button written inside the form; a form with neither has no way to be
	// sent.
	HideSubmit bool
}
```

## Example

```go
{{#include ../../../demos/form.go:demo}}
```

An untouched field reads as its `Default`, so the fields hold zero until the
app user fills them in and hits Submit — there is nothing to nil-check.

### Submitting from a button inside the form

```go
{{#include ../../../demos/form.go:button}}
```

`searched` is true on the run the click was sent with, and `keyword` already
holds what was typed before it.

### Any widget belongs in a form

A slider or a toggle inside a form holds its value like the rest of them, and
hands it over on submit rather than on the drag or the click:

```go
{{#include ../../../demos/form.go:widget}}
```

### Naming the submit button

```go
{{#include ../../../demos/form.go:label}}
```

<div data-toolgui-demo="form" data-toolgui-demo-height="800"></div>
