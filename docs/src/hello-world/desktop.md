# Desktop App

The same App can run in a desktop window instead of a browser, through
[Wails v2](https://wails.io). Only the executor changes: pages, components and
state work exactly as they do on the web.

```go
package main

import (
	"log"

	"github.com/voilelab/toolgui/toolgui/tgcomp"
	"github.com/voilelab/toolgui/toolgui/tgframe"

	tgwails "github.com/voilelab/toolgui/toolgui-wails"
)

func main() {
	app := tgframe.NewApp()
	app.AddPage("index", "Index", func(p *tgframe.Params) error {
		tgcomp.Text(p.Main, "Hello world")
		return nil
	})

	e := tgwails.NewExecutor(app, &tgwails.Conf{Title: "ToolGUI Hello"})
	if err := e.Run(); err != nil {
		log.Fatal(err)
	}
}
```

`Run` opens the window and blocks until the user closes it.

## The example app

[`toolgui-wails/example/hello`](https://github.com/voilelab/toolgui/tree/dev/toolgui-wails/example/hello)
is a runnable version of that program, with a sidebar textbox and a button to
show the state round trip working in a window:

```go
func Main(p *tgframe.Params) error {
	name := tgcomp.Textbox(p.State, p.Sidebar, "What's your name?")
	if name != "" {
		tgcomp.Text(p.Sidebar, "Hi "+name+"~")
	}

	tgcomp.Text(p.Main, "hello ")
	if tgcomp.Button(p.State, p.Main, "keep going") {
		tgcomp.Text(p.Main, "world")
	}

	return nil
}

func main() {
	app := tgframe.NewApp()
	app.AddPage("index", "Index", Main)

	e := tgwails.NewExecutor(app, &tgwails.Conf{Title: "ToolGUI Hello"})
	err := e.Run()
	if err != nil {
		log.Fatal(err)
	}
}
```

The whole app is two files:

```text
toolgui-wails/example/hello/
├── main.go      the app
└── wails.json   what the wails CLI builds
```

```json
{
  "$schema": "https://wails.io/schemas/config.v2.json",
  "name": "hello",
  "outputfilename": "hello",

  "frontend:dir": "../../../toolgui-web/wails",
  "frontend:install": "yarn",
  "frontend:build": "yarn build",

  "assetdir": "../../frontend/dist",
  "wailsjsdir": "./build"
}
```

* `frontend:dir` is the desktop frontend workspace. The CLI runs yarn there,
  which builds into `toolgui-wails/frontend/dist`.
* `assetdir` is that same directory, the one `assets.go` embeds and Wails
  serves the window from.
* `wailsjsdir` is where the generated JavaScript bindings land. The example
  does not import them — the frontend calls the bound methods through
  `window.go` — so they are build output, not source.

## Running it

Building a desktop app needs Go, Node with yarn for the frontend build, and
the webview development packages. On Debian and Ubuntu:

```shell
sudo apt-get install libgtk-3-dev libwebkit2gtk-4.1-dev
```

Then, from the repository root:

```shell
task run_wails_hello     # dev mode: frontend from disk, Go files watched
task build_wails_hello   # packaged binary, in example/hello/build/bin
```

The wails CLI itself is pinned as a tool dependency of the `toolgui-wails`
module, so there is nothing to install: both tasks run `go tool wails` in
`example/hello`. They also stub the embedded assets first, because the CLI
generates the bindings — which compiles the package holding the `//go:embed` —
before it builds the frontend the embed points at.

### Build tags

* `webkit2_41` asks for WebKit2GTK 4.1, on Linux only. The tasks add it there;
  pass `TAGS=` to drop it on a distribution that still ships 4.0, which is what
  Wails asks for by default.
* `production` picks the real app over a stub that refuses to run. The CLI adds
  it, and `dev` in dev mode.

A plain `go build` works too, it just has to supply what the CLI would. See
[toolgui-wails/README.md](https://github.com/voilelab/toolgui/blob/dev/toolgui-wails/README.md)
for that and the other platform notes.

## Conf

```go
type Conf struct {
	// Title shows in the title bar.
	Title string

	// Width and Height are the window size in pixels.
	Width, Height int

	// Background is the window background colour.
	Background *RGBA

	// DisableResize stops the user resizing the window.
	DisableResize bool

	// Frameless drops the window decorations.
	Frameless bool
}
```

Empty fields fall back to `DefaultConf`: a 1024x768 white window titled
`ToolGUI`.

## A separate module

`github.com/voilelab/toolgui/toolgui-wails` is its own Go module, because Wails
needs cgo, GTK and WebKit. Apps that only target the web never pull any of that
in.
