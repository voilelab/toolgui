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
		{Src: "/assets/icon.png", Type: "image/png", Sizes: "512x512"},
	},
	Display:         "standalone",
	ThemeColor:      "#000000",
	BackgroundColor: "#ffffff",
})
e.StartService(":3001")
```

toolgui ships no icon of its own, so an app that wants one serves the file
itself: see [Serving your own files](#serving-your-own-files) below. Empty
fields are left out of the served json, so the manifest holds only what the
app sets.

`SetManifest(nil)` goes back to the default, which an app that never calls it
serves too: `tgexec.DefaultManifest()`, named after the
[app title](index.md#title) when it has one.

## Serving your own files

`SetAssets` puts an `fs.FS` under `/assets/`, which is where a manifest icon,
or any other file the app hands the browser, comes from:

```go
//go:embed assets
var assets embed.FS

func main() {
	// ...

	e := tgexec.NewWebExecutor(app)

	// assets/icon.png is served at /assets/icon.png.
	sub, _ := fs.Sub(assets, "assets")
	e.SetAssets(sub)

	e.SetManifest(&tgexec.Manifest{
		Name: "My Tool",
		Icons: []tgexec.ManifestIcon{
			{Src: "/assets/icon.png", Type: "image/png", Sizes: "512x512"},
		},
	})

	e.StartService(":3001")
}
```

The files at the root of the fs are what `/assets/` shows, so `os.DirFS` works
the same way. Without a `SetAssets` call, `/assets/` holds nothing.

Both setters read their value per request, and both are safe to call while the
server is already serving, so an app can swap its manifest or its files at any
point in its own run.

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
