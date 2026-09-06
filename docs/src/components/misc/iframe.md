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
| `Height` | CSS height of the iframe.                         | `150px`   |
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

Show a button that sends a value back to the server.

In the iframe, there are four values injected on `window`:

* `window.update(value)` - function. Send an arbitrary JSON value back to the
  server. It is stored in the state under the **iframe's own id** — the iframe
  cannot write to another component's state.
* `window.upload(file)` - function. Upload a `File` to the server. Returns a
  `Promise<{ok: boolean, error?: string}>`.
* `window.theme` - current theme.
* `window.props` - props of the iframe component.

They are injected once the iframe document has loaded, and re-injected on every
reload, so a script that runs while the document is parsing should not touch
them at top level.

```go
tgcomp.IframeWithConf(
	p.Main,
	`<button id="btn">Click me to update</button>
	<script>
		const btn = document.getElementById('btn');
		btn.addEventListener('click', (event) => {
			window.update({clicked: true});
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
