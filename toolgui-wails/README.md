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

In-repo development builds against the working tree:

```
replace github.com/voilelab/toolgui => ../
```

A replace only applies in the main module, so it does not follow the module to
anyone depending on it. What they resolve is the `require` above it, which
`release.yml` stamps with the version being tagged.

## Installing it

```shell
go get github.com/voilelab/toolgui/toolgui-wails
```

Being a module in a subdirectory, it is tagged `toolgui-wails/vX.Y.Z` rather
than `vX.Y.Z`; `go get` finds that on its own. Desktop builds still need the
webview dependencies below.

## What crosses the boundary

| Web | Desktop |
| --- | --- |
| `GET /api/app` | `window.go.tgwails.ToolGUI.AppConf()` |
| update websocket | `Update(eventJSON)` + the `toolgui:pack` event |
| `POST /api/files` | `UploadFileStart` + `UploadFileChunk` + `UploadFileFinish` |
| a page load | `Start(pageName)` |
| the menubar the frontend draws | the window's own, from `options.App.Menu` |

Payloads cross as JSON strings — the same ones the websocket carries, so both
transports share a wire format. An upload is the exception: bound methods take
strings, so a file crosses in chunks rather than as one base64 blob. The state
writes them to disk as they land, and the component sees the file only once
the last chunk is in. Packs go out on a single event name so
create/update/delete/result keep the order the page produced them in.

## Links

A desktop window has no tabs and no back button, so letting an external link
navigate would replace the running app, page state and all. The adapter
catches clicks on the document and hands `http(s)` urls whose origin is not
the app's own to `runtime.BrowserOpenURL`, which opens them in the system
browser. Page navigation and other schemes are left alone, and `tccontent`
renders the same plain anchors it does on the web — nothing to set on `Link`
or `LinkButton`.

Wails serves the frontend from its own origin, so the asset server could carry
the plain HTTP endpoints too. Bound methods handle all four instead, to keep
this module off `tgexec`, which embeds the whole web bundle.

## Menus

`App.SetMenu` declares one tree, and the desktop window draws it itself:

```go
app.SetMenu(tgframe.NewMenu().
	Submenu("File", func(m *tgframe.Menu) {
		m.Text("Open", "file_open")
		m.Separator()
		m.Text("Quit", "file_quit")
	}))
```

The same declaration is a row of buttons above the app in a browser. Here it
is translated into a `menu.Menu` and handed to `options.App.Menu`, so what the
window shows is the platform's real menubar. The tree is then kept out of the
app config the frontend receives — otherwise it would draw its own menubar
inside the window, under the real one.

Nothing new carries the click back. A Wails `MenuItem.Click` callback happens
in Go, so it goes into the running session as the click event the web menubar
would have sent over the wire:

```go
session.HandleEvent(&tgframe.EventClick{ID: "menu_item_file_open"})
```

which is the ordinary rerun, with the packs going out to the frontend as
usual. The page reads the pick with `tgframe.MenuClicked(p, "file_open")`
whichever executor it is running under. The wire format does not change and
there is no new event type.

A pick made before `Start` has opened a session — the window is up, the
menubar with it, and no page has loaded yet — has nothing to run and is
ignored.

### What the platform adds

macOS actions like Cmd+C, Cmd+V and Cmd+Q belong to the standard App and Edit
menus rather than to the webview, so a window that replaces the menubar
without them loses those shortcuts. The App and Edit menus therefore go in
front of the app's own entries there, plus the Window menu unless the window
is frameless — the same set, under the same rule, that Wails gives a window
with no menu at all. So an app keeps the standard menus whether or not it
declares one of its own.

This is the one place the menubar is not the same on every platform, and it
cannot be helped: Wails v2 serves menu roles on macOS only. The translation
does not expose roles at all — a `tgframe.Menu` is text items, separators and
submenus, and nothing else.

### What it does not do

- **A menu item cannot change page.** On the desktop the page switch happens
  in the frontend, and Go has no way to ask it for one: `SendPackFunc` carries
  create, update, delete, ready and result, none of which navigate. That would
  need a new Go-to-frontend pack type.
- **Nothing is enabled or disabled while running.** The tree is declared once
  on the app, which is the shape a native menu — with no diff to apply — can
  take. `runtime.MenuUpdateApplicationMenu` rebuilds the whole menu, closing
  any dropdown that happens to be open, so a menu that changed on every run
  would be a menu that cannot be used.

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
