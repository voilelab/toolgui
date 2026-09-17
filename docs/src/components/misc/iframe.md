# Iframe (Experimental)

Iframe is used to show a html in a iframe.

> Warning: This component is experimental 🧪 and may not work as expected.

The iframe is sandboxed on an opaque origin, so the html cannot reach the app's
page, its DOM, its cookies or its storage. It talks to the app only through
`window.toolgui`, and can only write to its own state.

## API

```go
func Iframe(c *tgframe.Container, html string, conf ...*IframeConf)
func IframeValue[T any](c *tgframe.Container, html string, conf ...*IframeConf) *T
```

* `c` is the container to add the iframe to.
* `html` is the html to show in the iframe.
* `conf` is an optional configuration, at most one.

`IframeConf`:

| Field    | Description                                       | Default   |
| -------- | ------------------------------------------------- | --------- |
| `ID`     | The user specific id, from the embedded `tgframe.Base`. | hashed id |
| `Script` | Allow the iframe to run javascript.               | `false`   |
| `Width`  | CSS width of the iframe.                          | `100%`    |
| `Height` | CSS height of the iframe, or `auto`.              | `150px`   |

`IframeValue` returns the latest value the guest sent through
`window.toolgui.update`. It reads the value, it does not draw the iframe: give
it the same `html` and `conf` the `Iframe` call gets, so that both name the same
component. It returns `nil` until the guest has a value, which is not the same
as a value that happens to be the zero `T`. A guest that has not sent yet and
one that sent `null` both read as `nil` — `null` is how a guest says it has
nothing. It fails the run rather than return a silent zero when what the guest
sent does not fit `T`.

## Examples

### Simple

Show h1 element in the iframe.

```go
{{#include ../../../demos/iframe.go:simple}}
```

### Script

Run a script inside the iframe to update its content.

```go
{{#include ../../../demos/iframe.go:script}}
```

### Sizing

```go
tgcomp.Iframe(p.Main, "<h1>Hello World</h1>", &tgcomp.IframeConf{
	Script: true,
	Width:  "300px",
	Height: "400px",
})
```

### Interactive

`window.toolgui` is available inside every iframe that has `Script` enabled. It
talks to the app over `postMessage`, which is what lets the iframe stay on an
opaque origin.

* `window.toolgui.update(value)` - send an arbitrary JSON value back to the
  server. It is stored in the state under the **iframe's own id** — the app
  fills the id in, so an iframe cannot write to another component's state.
* `window.toolgui.upload(file)` - upload a `File`. Returns a
  `Promise<{ok: boolean, error?: string}>`.
* `window.toolgui.onRender(fn)` - run `fn(props, theme)` on every render, and
  once immediately if a render already arrived.
* `window.toolgui.autoHeight()` - report the document height to the app. Pair
  it with `Height: "auto"`.
* `window.toolgui.props` / `.theme` / `.id` - the latest values from the app.
  Undefined until the first render.

```go
{{#include ../../../demos/iframe.go:value}}
```

```go
{{#include ../../../demos/iframe.go:interactive}}
```

### Reacting to reruns, and auto height

`onRender` fires on the first render and again whenever the props or the theme
change, so an iframe can update itself without being reloaded.

`autoHeight` reports the guest's own height as it changes; with `Height: "auto"`
the app resizes the iframe to match.

Two caveats. It measures the body, so a guest whose body is sized off the
viewport (`height: 100%`) will feed its own height back and should set a fixed
`Height` instead. And because the iframe is isolated it is a separate rendering
context, which the browser throttles while it is scrolled out of view: an
offscreen iframe stays at the default height and takes its real one as it
becomes visible.

```go
{{#include ../../../demos/iframe.go:render}}
```
