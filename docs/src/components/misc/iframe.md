# Iframe (Experimental)

Iframe is used to show a html in a iframe.

> Warning: This component is experimental 🧪 and may not work as expected.

> **Security warning**: an iframe created with `script` enabled is **not**
> isolated from the app. It runs with `allow-scripts allow-same-origin`, which
> means the html shares the app's origin and can reach the parent page, its
> DOM and its storage. Only pass html you trust as much as your own app code.
> A redesign is tracked in
> [issue #30](https://github.com/voilelab/toolgui/issues/30).

## API

```go
func Iframe(c *tgframe.Container, html string, script bool)
func IframeWithID(c *tgframe.Container, html string, script bool, id string)
func IframeWithConf(c *tgframe.Container, html string, conf *IframeConf)
func IframeValue(s *tgframe.State, id string, out any) error
```

* `c` is the container to add the iframe to.
* `html` is the html to show in the iframe.
* `script` is used to allow the iframe to run javascript. (notice that this is not secure)
* `id` is the user specific id.

`IframeConf`:

| Field    | Description                                       | Default   |
| -------- | ------------------------------------------------- | --------- |
| `Script` | Allow the iframe to run javascript.               | `false`   |
| `Width`  | CSS width of the iframe.                          | `100%`    |
| `Height` | CSS height of the iframe, or `auto`.              | `150px`   |
| `ID`     | The user specific id.                             | hashed id |

`IframeValue` unmarshals the latest value the iframe sent through
`window.update` into `out`. The `id` is the one passed to `IframeWithID` or
`IframeConf.ID`, so an interactive iframe needs an explicit id — the default id
is derived from a hash of the html.

## Examples

### Simple

Show h1 element in the iframe.

```go
tgcomp.Iframe(p.Main, "<h1>Hello World</h1>", true)
```

### Script

Run a script inside the iframe to update its content.

```go
htmlWithScript := `
<b id="test">Hello world not changed</b>
<script>
	const element = document.getElementById('test');
	element.innerText = 'Hello world gen by script';
</script>`

tgcomp.IframeWithID(
	p.Main,
	htmlWithScript,
	true,
	"iframe_with_script")
```

### Sizing

```go
tgcomp.IframeWithConf(p.Main, "<h1>Hello World</h1>", &tgcomp.IframeConf{
	Script: true,
	Width:  "300px",
	Height: "400px",
})
```

### Interactive

`window.toolgui` is available inside every iframe that has `Script` enabled. It
talks to the app over `postMessage`, so it keeps working when the iframe is
isolated from the app.

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
tgcomp.IframeWithConf(
	p.Main,
	`<button id="btn">Click me to update</button>
	<script>
		const btn = document.getElementById('btn');
		btn.addEventListener('click', (event) => {
			window.toolgui.update({clicked: true});
		});
	</script>`,
	&tgcomp.IframeConf{
		Script: true,
		Height: "60px",
		ID:     "iframe_with_interactive",
	})

var value struct {
	Clicked bool `json:"clicked"`
}
err := tgcomp.IframeValue(p.State, "iframe_with_interactive", &value)
if err != nil {
	return err
}

tgcomp.Text(p.Main, fmt.Sprintf("Status: %v", value.Clicked))
```

### Reacting to reruns, and auto height

`onRender` fires on the first render and again whenever the props or the theme
change, so an iframe can update itself without being reloaded.

`autoHeight` reports the guest's own height as it changes; with `Height: "auto"`
the app resizes the iframe to match. It measures the body, so a guest whose body
is sized off the viewport (`height: 100%`) will feed its own height back and
should set a fixed `Height` instead.

```go
tgcomp.IframeWithConf(
	p.Main,
	`<div id="out">waiting for render</div>
	<script>
		const out = document.getElementById('out');
		window.toolgui.onRender((props, theme) => {
			out.innerText = 'theme=' + theme;
		});
		window.toolgui.autoHeight();
	</script>`,
	&tgcomp.IframeConf{
		Script: true,
		Height: "auto",
		ID:     "iframe_with_render",
	})
```

### Deprecated: the injected globals

`window.update`, `window.upload`, `window.props` and `window.theme` are still
injected and still work, but they only work because the iframe shares the app's
origin. They go away together with `allow-same-origin`, so port to
`window.toolgui`. They also have no way to tell the guest that a value changed,
which is what `onRender` is for.

