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

## A separate module

`github.com/voilelab/toolgui/toolgui-wails` is its own Go module, because Wails
needs cgo, GTK and WebKit. Apps that only target the web never pull any of that
in.

Building a desktop app needs the webview dependencies and the wails CLI. See
[toolgui-wails/README.md](https://github.com/voilelab/toolgui/blob/main/toolgui-wails/README.md)
for the build steps and platform notes.
