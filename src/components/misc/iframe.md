# Iframe (Experimental)

Iframe is used to show a html in a iframe.

> Warning: This component is experimental 🧪 and may not work as expected.

## API

```go
func Iframe(c *tgframe.Container, html string, script bool)
func IframeWithID(c *tgframe.Container, html string, script bool, id string)
```

* `c` is the container to add the iframe to.
* `html` is the html to show in the iframe.
* `script` is used to allow the iframe to run javascript. (notice that this is not secure)
* `id` is the user specific id.

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

### Interactive

Show a button that updates the iframe content.

In the iframe, there are four values injected on `window`:

* `window.update(data)` - function. Send data back to the server.
* `window.upload(data)` - function. Upload a file to the server.
* `window.theme` - current theme.
* `window.props` - props of the iframe component.

```go
tgcomp.IframeWithID(
	p.Main,
	`<button id="btn">Click me to update</button>
	<script>
		const btn = document.getElementById('btn');
		btn.addEventListener('click', (event) => {
			window.update({type: "click", id: "iframe_with_interactive_btn"});
		});
	</script>`,
	true,
	"iframe_with_interactive")
clickStatus := p.State.GetClickID() == "iframe_with_interactive_btn"
tgcomp.Text(p.Main, fmt.Sprintf("Status: %v", clickStatus))
```

