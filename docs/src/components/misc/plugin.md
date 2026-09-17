# Plugin (Experimental)

Plugin runs a script the app serves in a sandboxed frame, and hands it props
from Go.

> Warning: This component is experimental 🧪 and may not work as expected.

It is [`Iframe`](iframe.md) with the guest document turned inside out: instead
of a string of html in the page function, the guest is a script, a stylesheet
and whatever else the plugin is made of, shipped as files. The frame is the
same one — an opaque origin, no reach into the app's page, dom, cookies or
storage, and the `window.toolgui` bridge for everything it does need.

## API

```go
func (app *tgframe.App) AddPluginAssets(name string, fsys fs.FS) error
func tgframe.PluginAssetURL(name, file string) string

func Plugin(c *tgframe.Container, src string, conf ...*PluginConf)
func PluginValue[T any](c *tgframe.Container, src string, conf ...*PluginConf) *T
```

* `c` is the container to add the plugin to.
* `src` is the url of the plugin script.
* `conf` is an optional configuration, at most one.

`PluginValue` reads the plugin's value, it does not draw the plugin: give it
the same `src` and `conf` the `Plugin` call gets. It returns `nil` until the
plugin has a value, which is not the same as a value that happens to be the
zero `T`. A plugin that has not sent yet and one that sent `null` both read as
`nil` — `null` is how a plugin says it has nothing. It fails the run rather than
return a silent zero when what the plugin sent does not fit `T`.

`PluginConf`:

| Field    | Description                                          | Default  |
| -------- | ---------------------------------------------------- | -------- |
| `ID`     | Names the plugin's state, from the embedded `tgframe.Base`. | none |
| `Props`  | Props handed to the plugin.                          | `nil`    |
| `Style`  | Url of a stylesheet to load in the frame.            | none     |
| `Width`  | CSS width of the frame.                              | `100%`   |
| `Height` | CSS height of the frame, or `auto`.                  | `150px`  |

`ID` is the key `PluginValue` reads, and the id the frontend stamps on every
value the plugin sends. A plugin with no `ID` derives one from its `src`, so it
can still send values back; two plugins running the same script need an `ID`
each.

## Shipping a plugin

`AddPluginAssets` serves a file set under `/plugin/<name>/`, and
`PluginAssetURL` names a file in it. The name is one url path segment, so it
cannot reach outside the prefix it is served under.

```go
//go:embed plugins/colorpicker
var colorPickerAssets embed.FS

func main() {
	app := tgframe.NewApp()
	// ...

	assets, err := fs.Sub(colorPickerAssets, "plugins/colorpicker")
	if err != nil {
		log.Fatal(err)
	}

	if err := app.AddPluginAssets("colorpicker", assets); err != nil {
		log.Fatal(err)
	}

	tgexec.NewWebExecutor(app).StartService("127.0.0.1:3000")
}
```

The web and desktop executors both serve them, so a plugin works the same over
http and in a desktop window. The wasm build is the exception: it ships as
static files with no executor behind them, so nothing answers the plugin's url
there.

## Writing a plugin

The script runs inside the frame, after the bridge is installed and after the
document is parsed, so `window.toolgui` and `document.body` are both there.

```js
(function () {
  var root = document.createElement('div')
  document.body.appendChild(root)

  window.toolgui.onRender(function (props, theme) {
    root.innerHTML = ''
    // ... draw props ...
  })

  window.toolgui.autoHeight()
})()
```

`onRender` fires on the first render and again whenever the props or the theme
change. `update` sends a value back to the app, which lands in the state under
the plugin's own id and is read with `PluginValue`. `autoHeight` reports the
plugin's height, which is what `Height: "auto"` follows. They are the same
bridge the [iframe](iframe.md) page documents.

It is worth keeping the plugin's own state in Go rather than in the frame: a
plugin that draws what its props say shows the same thing after a rerun or a
reconnect.

## Example

The demo's colour picker takes the colours and the current selection from Go,
and sends back the one the user clicks.

The selection is what the plugin draws from, so it is read before the plugin
is drawn.

```go
{{#include ../../../demos/plugin.go:value}}
```

```go
{{#include ../../../demos/plugin.go:demo}}
```

```js
(function () {
  var root = document.createElement('div')
  document.body.appendChild(root)

  window.toolgui.onRender(function (props, theme) {
    document.body.className = theme === 'dark' ? 'dark' : ''
    root.innerHTML = ''

    var colors = props.colors || []
    for (var i = 0; i < colors.length; i++) {
      root.appendChild(swatch(colors[i], colors[i] === props.selected))
    }
  })

  function swatch(color, selected) {
    var button = document.createElement('button')
    button.className = selected ? 'swatch selected' : 'swatch'
    button.style.background = color
    button.addEventListener('click', function () {
      window.toolgui.update({ color: color })
    })
    return button
  }

  window.toolgui.autoHeight()
})()
```

## Limits

The frame is the tradeoff. A plugin gets none of the app's css or theme beyond
what `onRender` hands it, it cannot contain other toolgui components, and each
one is a document of its own.

The other limit is where the app runs: a plugin is loaded over a url, so it
needs an executor serving one. That rules out the wasm build, which is a
directory of static files.
