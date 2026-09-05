# toolgui-wails

Runs a ToolGUI app in a desktop window with [Wails v2](https://wails.io),
instead of serving it over HTTP.

```go
app := tgframe.NewApp()
app.AddPage("index", "Index", Main)

e := tgwails.NewExecutor(app, &tgwails.Conf{Title: "My Tool"})
e.Run()
```

The pages, components and state are the same as the web executor's: only the
transport changes. `tgframe.Session` does the work either way, and the frontend
is the same `@toolgui-web/lib` components with a different adapter around them.

## Why a separate module

Wails needs cgo, GTK and WebKit. Keeping it in its own module means people who
only build web apps never pull any of that in.

The module uses the parent from the working tree:

```
replace github.com/voilelab/toolgui => ../
```

toolgui has no tags yet, so there is no released version to require instead.

## What crosses the boundary

| Web | Desktop |
| --- | --- |
| `GET /api/app` | `window.go.tgwails.ToolGUI.AppConf()` |
| update websocket | `Update(eventJSON)` + the `toolgui:pack` event |
| `POST /api/files` | `UploadFile(name, base64)` |
| a page load | `Start(pageName)` |

Payloads cross as JSON strings — the same ones the websocket carries, so both
transports share a wire format. Packs go out on a single event name so
create/update/delete/result keep the order the page produced them in.

Wails serves the frontend from its own origin, so the asset server could carry
the plain HTTP endpoints too. Bound methods handle all four instead, to keep
this module off `tgexec`, which embeds the whole web bundle.

## Building

The frontend is an ordinary Vite build that `assets.go` embeds:

```shell
task asset_lib
task asset_wails
```

Then, from this directory:

```shell
go run -tags "production,webkit2_41" ./example/hello
```

### Build tags

- `production` picks the real app. Without it Wails compiles a stub that
  returns `Wails applications will not build without the correct build tags.`
  Use `dev` instead for a debug build with devtools.
- `webkit2_41` asks for WebKit2GTK 4.1. Drop it on a distribution that still
  ships 4.0, which is what Wails asks for by default.

### Webview dependencies

On Debian and Ubuntu:

```shell
sudo apt-get install libgtk-3-dev libwebkit2gtk-4.1-dev
```

`.github/workflows/wails.yml` does the same and builds with both tags.
