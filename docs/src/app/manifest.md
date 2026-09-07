# Web Manifest

A browser reads `/manifest.json` to name the app, pick its icons, and decide
how it looks when it is installed to a home screen. `WebExecutor` serves that
file, so an app configures it in Go:

```go
e := tgexec.NewWebExecutor(app)
e.SetManifest(&tgexec.Manifest{
	Name:      "My Tool",
	ShortName: "My Tool",
	Icons: []tgexec.ManifestIcon{
		{Src: "/static/icon.png", Type: "image/png", Sizes: "512x512"},
	},
	Display:         "standalone",
	ThemeColor:      "#000000",
	BackgroundColor: "#ffffff",
})
e.StartService(":3001")
```

toolgui ships no icon of its own, so an app that wants one serves the file and
points `Icons` at it. Empty fields are left out of the served json, so the
manifest holds only what the app sets.

`SetManifest(nil)` goes back to the default, which an app that never calls it
serves too: `tgexec.DefaultManifest()`, named after the
[app title](index.md#title) when it has one.

## Other members

`Extra` carries the manifest members the struct doesn't name, and a key there
wins over the field of the same name:

```go
e.SetManifest(&tgexec.Manifest{
	Name: "My Tool",
	Extra: map[string]any{
		"categories": []string{"utilities"},
	},
})
```

Only the desktop executor has no manifest: the [desktop
app](../hello-world/desktop.md) is not installed through a browser.
