# Custom Components

ToolGUI ships a fixed set of components, but a page function is ordinary Go,
so the set is not the limit. There are three ways to add something of your
own, and they cost very different amounts of work. Start at the top.

## A function over the built-ins

Most of what people call a custom component is a fixed arrangement of
components that already exist. That is a function taking a container:

```go
// Metric shows a label with the value under it.
func Metric(c *tgframe.Container, label string, value float64) {
	box := tgcomp.Box(c, "metric_"+label)
	tgcomp.Text(box, label)
	tgcomp.Title(box, strconv.FormatFloat(value, 'f', 2, 64))
}
```

Nothing registers it and nothing knows it exists: by the time the run
reaches the frontend it is a box with two components in it.

Two things to watch for.

**Ids have to stay unique.** Everything stateful inside your function claims
an id, and two calls on one page claim it twice, which fails the run with
`duplicated component id`. Take the id as an argument and pass it down, the
way the built-in components do with `Conf.ID`:

```go
func Metric(c *tgframe.Container, id, label string, value float64) {
	box := tgcomp.Box(c, id)
	// ...
}
```

Components that hold no state — `Text`, `Title`, `Markdown` — have no id
unless you give them one, so a display-only function needs no id argument at
all.

**Reading a value back** works the way input components work: the state is
keyed by id, so read it and return it.

```go
// Counter shows a number and a button that adds one to it.
func Counter(s *tgframe.State, c *tgframe.Container, id string) int {
	count := 0
	if v := s.GetInt(id); v != nil {
		count = *v
	}

	if tgcomp.ButtonWithConf(s, c, "+1", &tgcomp.ButtonConf{ID: id + "_button"}) {
		count++
		s.Set(id, count)
	}

	tgcomp.Text(c, strconv.Itoa(count))
	return count
}
```

## A component struct

A component is anything that can tell you its id:

```go
type Component interface {
	GetID() string
}
```

Embed [`tgframe.BaseComponent`](https://pkg.go.dev/github.com/voilelab/toolgui/toolgui/tgframe#BaseComponent)
to get that for free, and every exported field becomes a prop: the component
is marshalled to JSON as it is and sent to the frontend.

```go
type sparklineComponent struct {
	*tgframe.BaseComponent
	Points []float64 `json:"points"`
}

func Sparkline(c *tgframe.Container, points []float64) {
	c.AddComponent(&sparklineComponent{
		BaseComponent: &tgframe.BaseComponent{
			Name: "sparkline_component",
		},
		Points: points,
	})
}
```

`Name` is the type the frontend renders by. `ID` is the key its state is
stored under, and is only needed by a component that has state — see
[Components](components.md) for how the two differ from the position key the
container assigns.

The catch is the other half. `Name` is looked up in a map that is fixed when
the web assets are built, and those assets are embedded into the Go binary,
so a name nothing renders does not draw a blank — it fails the page. Adding
a renderer means building your own frontend, which is a real fork of the
project, not an extension of it.

## Riding on a component that already renders

The way to ship a component with its own rendering, without forking the
frontend, is to build it out of one that the frontend already knows.

[`Iframe`](../components/misc/iframe.md) is the one meant for it. It takes
HTML, runs it sandboxed on an opaque origin, and gives the guest a
`window.toolgui` bridge: `onRender` for props and theme, `update` to write a
value back into the state, `upload` for files, `autoHeight` for sizing.

Wrap it, and the iframe becomes an implementation detail your callers never
see:

```go
// Gauge draws a dial, and reports the value the user leaves it on.
func Gauge(s *tgframe.State, c *tgframe.Container, id string, value float64) float64 {
	tgcomp.IframeWithConf(c, gaugeHTML, &tgcomp.IframeConf{
		Script: true,
		Height: "auto",
		ID:     id,
	})

	var out struct {
		Value float64 `json:"value"`
	}
	if err := tgcomp.IframeValue(s, id, &out); err != nil {
		return value
	}

	return out.Value
}
```

An interactive iframe needs an explicit `ID`: without one its id is a hash of
the HTML, which is not something a caller can name.

[`Plugin`](../components/misc/plugin.md) is the same frame with the guest
shipped as files instead of a string. Register the files on the app and name
them by url:

```go
//go:embed plugins/gauge
var gaugeAssets embed.FS

assets, _ := fs.Sub(gaugeAssets, "plugins/gauge")
app.AddPluginAssets("gauge", assets)
```

```go
tgcomp.Plugin(c, id, tgframe.PluginAssetURL("gauge", "gauge.js"), props)
```

The props are whatever you hand it, marshalled to json and delivered to
`onRender`; the value the plugin sends back is read with `PluginValue`. Reach
for it over `Iframe` as soon as the guest is more than a few lines: it is a
javascript file with an editor and a linter around it, rather than a string
in a Go file.

The tradeoffs are the frame's, either way. The guest cannot reach the app's
DOM, cookies or storage, which is the point, but it also gets none of the
app's CSS or theme beyond what `onRender` hands it, it cannot contain other
toolgui components, and each instance is a document of its own.
