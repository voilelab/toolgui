# toolgui-wails

Runs a ToolGUI app in a desktop window with [Wails v1](https://github.com/wailsapp/wails),
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

Wails v1 needs cgo, GTK and WebKit. Keeping it in its own module means people
who only build web apps never pull any of that in.

The module uses the parent from the working tree:

```
replace github.com/voilelab/toolgui => ../
```

toolgui has no tags yet, so there is no released version to require instead.

## What crosses the boundary

| Web | Desktop |
| --- | --- |
| `GET /api/app` | `window.backend.ToolGUI.AppConf()` |
| update websocket | `Update(eventJSON)` + the `toolgui:pack` event |
| `POST /api/files` | `UploadFile(name, base64)` |
| a page load | `Start(pageName)` |

Wails converts a JSON argument straight to the parameter type, which rules out
structs, so every payload crosses as a string. Packs go out on a single event
name so create/update/delete/result keep the order the page produced them in.

## Building

The frontend is bundled into one JS and one CSS file, which `assets.go` embeds,
because that is what Wails v1 injects into the webview:

```shell
task asset_lib
task asset_wails
```

Then, from this directory:

```shell
go run ./example/hello
```

### Webview dependencies

Building needs GTK 3 and WebKit2GTK. On Debian and Ubuntu:

```shell
sudo apt-get install libgtk-3-dev libwebkit2gtk-4.0-dev
```

Ubuntu dropped `libwebkit2gtk-4.0-dev` after 22.04, while wails v1 still asks
pkg-config for `webkit2gtk-4.0`. On a newer release, install
`libwebkit2gtk-4.1-dev` and point the old name at it:

```shell
printf 'Name: webkit2gtk-4.0 shim\nDescription: shim\nVersion: %s\nRequires: webkit2gtk-4.1\n' \
  "$(pkg-config --modversion webkit2gtk-4.1)" \
  | sudo tee /usr/lib/x86_64-linux-gnu/pkgconfig/webkit2gtk-4.0.pc
```

`.github/workflows/wails.yml` does the same thing.
