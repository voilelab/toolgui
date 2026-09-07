# Browser App (WebAssembly)

The same App can run *inside* the browser, compiled to WebAssembly, with no Go
process behind it. Only the executor changes: pages, components and state work
exactly as they do on the web.

```go
//go:build js && wasm

package main

import (
	"github.com/voilelab/toolgui/toolgui/tgcomp"
	"github.com/voilelab/toolgui/toolgui/tgframe"
	"github.com/voilelab/toolgui/toolgui/tgwasm"
)

func main() {
	app := tgframe.NewApp()
	app.AddPage("index", "Index", func(p *tgframe.Params) error {
		tgcomp.Text(p.Main, "Hello world")
		return nil
	})

	// A static host cannot route paths, so pages live in the hash.
	app.SetHashPageNameMode(true)

	tgwasm.NewExecutor(app).Run()
}
```

`Run` installs the bridge the page talks to and blocks forever, keeping the
wasm instance alive to answer it.

`tgwasm` is behind `//go:build js && wasm`, so an app that also ships a server
build splits its entry points:

```text
page.go          the pages, no build tag
main_server.go   //go:build !(js && wasm)   tgexec.NewWebExecutor(app).StartService(":3000")
main_wasm.go     //go:build js && wasm      tgwasm.NewExecutor(app).Run()
```

## Building it

`toolgui-wasm` compiles the app and writes the site around it:

```shell
go get -tool github.com/voilelab/toolgui/cmd/toolgui-wasm

go tool toolgui-wasm build -o dist ./cmd/myapp   # the site, into dist/
go tool toolgui-wasm serve ./cmd/myapp           # the same, at :3000
```

`serve` is there so the browser gets the binary as `application/wasm` and can
compile it while it downloads; deploying needs no server at all. Neither
command deletes anything it did not write, so `-o` can point at a web root.

## The example app

[`toolgui/tgwasm/example/hello`](https://github.com/voilelab/toolgui/tree/dev/toolgui/tgwasm/example/hello)
is a runnable version, with three pages, a sidebar textbox, a button and a file
the page function reads without it leaving the tab:

```shell
task build_wasm_hello   # static site, in toolgui/tgwasm/example/hello/build
task run_wasm_hello     # the same, served at http://localhost:3000
```

The component demo runs in the browser too — `task run_wasm_demo`, and it is
[published beside this book](../demo/).

## What a build produces

```text
build/
├── index.html      the page
├── static/         the frontend bundle and the worker
├── wasm_exec.js    the Go runtime shim
└── app.wasm        your app
```

Four static files. Any file server serves them; there is no backend.

The CLI does nothing a shell cannot:

```shell
GOOS=js GOARCH=wasm go build -o dist/app.wasm ./your/package
cp "$(go env GOROOT)/lib/wasm/wasm_exec.js" dist/
# plus the frontend, which the CLI carries as an embedded copy
```

`wasm_exec.js` is copied out of the toolchain that built the binary rather than
vendored: the two have to match. That is also why the CLI shells out to `go
build` instead of asking you to run it — one toolchain, both halves.

## Where it runs

The Go program runs in a Web Worker, not on the page's thread. It never touches
the DOM — it sends the same packs the update websocket carries, and the React
components on the main thread render them — so a page function that takes a
while leaves the UI responsive.

## What the browser takes away

* One session per tab, created on load. A reload starts from an empty state:
  there is no server to keep it.
* No filesystem and no listening socket. `net/http` requests become `fetch`, so
  CORS applies to whatever your page function calls.
* `GOMAXPROCS` is 1. Goroutines interleave, nothing runs in parallel.
* No timezone database unless the app imports `time/tzdata`.
* Uploaded files are held in memory.
* The binary is public, like any other static asset. No secrets in it.
* `SetManifest` and `SetAssets` are `WebExecutor` settings, so the browser
  build does without them. A static site can carry a `manifest.json` and its
  own files next to `index.html` instead.

## Hosting notes

* Serve `.wasm` as `application/wasm`, so the browser can compile it while it
  downloads.
* Compress it. A small app is around 5 MB, about 1.4 MB gzipped; the component
  demo is 6.7 MB, 1.8 MB gzipped. It is cached after the first load.
* Keep `SetHashPageNameMode(true)` unless the host can rewrite unknown paths to
  `index.html`.
* The build uses relative asset URLs, so it works at a site root and under a
  project path like `/toolgui/` without rebuilding.
