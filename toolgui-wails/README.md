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

The wails CLI is pinned as a tool dependency of this module, so there is
nothing to install:

```shell
task run_wails_hello     # dev mode: frontend from disk, Go files watched
task build_wails_hello   # packaged binary, in example/hello/build/bin
```

Both run `go tool wails` in `example/hello`, whose `wails.json` points the CLI
at the frontend workspace. The CLI builds the frontend into `frontend/dist`,
generates the bindings and compiles the app, and supplies the build flags a
desktop build needs — including the macOS frameworks, which are easy to miss
by hand.

One ordering detail: the CLI generates bindings *before* it builds the
frontend, and generating them compiles this package, whose `assets.go` embeds
`frontend/dist`. So the embed has to resolve before any of it runs, which is
what `task stub_assets` is for. Both tasks do that first.

### Build tags

- `webkit2_41` asks for WebKit2GTK 4.1, on Linux only. The tasks add it there;
  pass `TAGS=` to drop it on a distribution that still ships 4.0, which is
  what Wails asks for by default.
- `production` picks the real app over a stub that refuses to run. The CLI
  adds it, and `dev` in dev mode.

### Without the CLI

The module itself is an ordinary Go package, so a plain build works — it just
has to supply what the CLI would:

```shell
# macOS 11+: wails calls UTType for its file dialogs but never links the
# framework that defines it, so the link fails on _OBJC_CLASS_$_UTType.
CGO_LDFLAGS="-framework UniformTypeIdentifiers" go build -tags production ./example/hello
```

### Webview dependencies

On Debian and Ubuntu:

```shell
sudo apt-get install libgtk-3-dev libwebkit2gtk-4.1-dev
```

`.github/workflows/wails.yml` does the same and builds with both tags.
