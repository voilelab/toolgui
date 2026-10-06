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

Pages are then linked as `/#/{name}` — `#/index` here, not `#index`. See
[Page name in the URL](../app/page.md#page-name-in-the-url).

`Run` installs the bridge the page talks to and blocks forever, keeping the
wasm instance alive to answer it.

`tgwasm` is behind `//go:build js && wasm`, so an app that also ships a server
build splits its entry points:

```text
page.go          the pages, no build tag
main_server.go   //go:build !(js && wasm)   tgexec.NewWebExecutor(app).StartService("127.0.0.1:3000")
main_wasm.go     //go:build js && wasm      tgwasm.NewExecutor(app).Run()
```

## Building it

`toolgui-wasm` compiles the app and writes the site around it:

```shell
go get -tool github.com/voilelab/toolgui/cmd/toolgui-wasm

go tool toolgui-wasm build -o dist ./cmd/myapp   # the site, into dist/
go tool toolgui-wasm serve ./cmd/myapp           # the same, at :3000
```

`-ldflags` goes through to `go build`, e.g. `-ldflags="-X main.version=1.0"`.
`-s -w` saves little here: wasm keeps most of its size in code, not symbols.

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
├── manifest.json   the web app manifest
├── sw.js           the service worker, with -offline
├── static/         the frontend bundle and the worker
├── wasm_exec.js    the Go runtime shim
└── app.wasm        your app
```

A handful of static files. Any file server serves them; there is no backend.
It does have to be an `https` one, though, or `localhost` — see the hosting
notes.

The CLI does nothing a shell cannot:

```shell
GOOS=js GOARCH=wasm go build -o dist/app.wasm ./your/package
cp "$(go env GOROOT)/lib/wasm/wasm_exec.js" dist/
# plus the frontend, which the CLI carries as an embedded copy
```

`wasm_exec.js` is copied out of the toolchain that built the binary rather than
vendored: the two have to match. That is also why the CLI shells out to `go
build` instead of asking you to run it — one toolchain, both halves.

## Web manifest

The browser reads `manifest.json` before any Go runs, so the app cannot set it
from `main`. `toolgui-wasm` writes it instead:

```shell
go tool toolgui-wasm build -manifest manifest.json -assets assets ./cmd/myapp
```

* `-manifest` is a json file, written as `manifest.json`. It takes the same
  members `tgexec.Manifest` does, `"icons"` and all.
* `-assets` is a directory, copied to `assets/`. `assets/icon.png` in it is
  `"src": "assets/icon.png"` in the manifest.
* Without `-manifest`, a build writes `tgexec.DefaultManifest()`, unless the
  output already has a `manifest.json` of its own, which it keeps.

```json
{
  "name": "My Tool",
  "short_name": "My Tool",
  "start_url": ".",
  "display": "standalone",
  "icons": [
    {"src": "assets/icon.png", "type": "image/png", "sizes": "512x512"}
  ]
}
```

Keep the urls relative, so the site still works under a project path.

## Icon

`-icon` points the favicon in `index.html` at a url, so the tab, bookmarks
and crawlers see it before any wasm loads. Ship the file with `-assets`:

```shell
go tool toolgui-wasm build -assets assets -icon assets/favicon.svg ./cmd/myapp
```

Once the app loads, the frontend sets the icon again from the app's config:
the page emoji, unless the app calls
[`app.SetIcon`](../app/index.md#icon). Set both to the same url to keep one
icon throughout.

## Head HTML

Crawlers run no wasm, so what they read has to be in `index.html` itself:
Open Graph or Twitter card meta, fonts, a CSP meta, analytics tags.
`App.SetTitle` runs too late for them. `-head` inserts a file at the
start of the head, right after the charset meta, so a CSP meta covers the links
after it:

```shell
go tool toolgui-wasm build -head head.html ./cmd/myapp
```

```html
<meta property="og:title" content="My Tool" />
<meta property="og:image" content="https://example.com/assets/og.png" />
```

The server executor does the same with
[`SetHeadHTML`](../app/manifest.md#head-html).

## Opening offline

`-offline` adds a service worker, so the app opens with no network once it has
been loaded once:

```shell
go tool toolgui-wasm build -offline -manifest manifest.json -o dist ./cmd/myapp
```

* The worker keeps a copy of the files the build wrote and serves it when the
  network fails. Online it gets out of the way: every load goes to the server
  and refreshes the copy, so a new build shows up on the next load, as it
  would without a worker.
* A build without `-offline` removes the worker on its first load, along with
  its copy.
* Only the site's own files are kept. What a page function fetches still needs
  the network, and the state is still the tab's: offline or not, a reload
  starts from an empty one.
* Service workers need a [secure context] too.

## Where it runs

The Go program runs in a dedicated Web Worker, not on the page's thread, and it
has to: the synchronous file handles the file store keeps uploads in exist in a
worker and nowhere else. It also never touches the DOM — it sends the same packs
the update websocket carries, and the React components on the main thread render
them — so a page function that takes a while leaves the UI responsive.

## Talking to the page

A worker's location is its script, not the page's URL, so what an app needs
to know about the page is handed over at boot:

* `tgwasm.Query()` is the page's query string, for an app that takes its own
  parameters such as `?lang=en`. Read once: it cannot change without a reload.
* `tgwasm.Embedded()` reports `?embed`, the frontend's display mode for a page
  shown in a frame, for an app laid out differently there.

The other way, `tgwasm.Emit(name, detail)` dispatches a `toolgui:<name>` DOM
event on the page's window, with `detail` sent as JSON. It reaches that tab
only, and is for a script of the page's own, such as analytics added with
[`-head`](#head-html):

```go
tgwasm.Emit("download", map[string]any{"tool": "novel"})
```

```js
addEventListener('toolgui:download', (e) => umami.track('download', e.detail))
```

`Emit` returns `tgwasm.ErrNoEventCallback` until the frontend has registered
for events, which it does at boot, before the first page runs.

## What the browser takes away

* One session per tab, created on load. A reload starts from an empty state:
  there is no server to keep it.
* No filesystem to open by path, and no listening socket. `net/http` requests
  become `fetch`, so CORS applies to whatever your page function calls.
* `GOMAXPROCS` is 1. Goroutines interleave, nothing runs in parallel.
* No timezone database unless the app imports `time/tzdata`.
* Uploaded files are kept in the origin private file system rather than the
  tab's memory, and go with the session. Storage there is per origin and
  bounded, so an upload can fail for want of room. Getting there is a stream:
  the worker copies the picked file straight into storage and hands Go a handle
  on it, so an upload is bounded by the origin's room for it and not by the
  tab's memory.
* The binary is public, like any other static asset. No secrets in it.
* `SetManifest` and `SetAssets` are `WebExecutor` settings. The browser
  build takes the same things as build flags instead; see
  [Web manifest](#web-manifest).

## Hosting notes

* Serve it over `https`. Uploads go to the origin private file system, and that
  belongs to a [secure context], so on plain `http` from anything but
  `localhost` there is nowhere to put one and every upload fails. It says so
  rather than falling back to the tab's memory, which would put the files back
  where this build stopped keeping them without anyone noticing. GitHub Pages,
  Netlify and the like are `https` already.
* Serve `.wasm` as `application/wasm`, so the browser can compile it while it
  downloads.
* Compress it. With Go 1.27 the example app is 8.2 MB, 2.2 MB gzipped; the
  component demo is 10.2 MB, 2.7 MB gzipped. It is cached after the first load,
  and the page shows how much of it has arrived until then.
* GitHub Pages gzips it for any browser that asks, which is all of them, so a
  visitor downloads the gzipped size. DevTools' Network panel shows both: the
  *transferred* size is what went over the wire, the *resource* size is the
  binary after decompression. `Content-Encoding: gzip` on the response is how to
  tell a host does it. A host that does not can serve a precompressed
  `app.wasm.gz` or `.br` instead, if it can set the header.
* Keep `SetHashPageNameMode(true)` unless the host can rewrite unknown paths to
  `index.html`.
* The build uses relative asset URLs, so it works at a site root and under a
  project path like `/toolgui/` without rebuilding.

## Why not TinyGo

TinyGo makes much smaller binaries, but it cannot build a ToolGUI app. Checked
with TinyGo 0.40.1:

* It supports Go up to 1.25, and this module needs 1.27.
* Past that, its standard library has no `encoding/json/v2` and no `uuid`,
  which `tgjson`, `tgframe` and `tgutil` import.

Worth another look when TinyGo catches up with the Go release.

[secure context]: https://developer.mozilla.org/en-US/docs/Web/Security/Secure_Contexts
