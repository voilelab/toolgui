# WebAssembly: survey for a browser-only build

How ToolGUI could ship an app as WebAssembly: the page function compiled to
`js/wasm` and run in the browser, with no Go process behind it. This picks a
shape and measures its cost.

**Status: phase 1 landed** — `toolgui/tgwasm`, the `toolgui-web/wasm` frontend
and the hello example are in the tree, built by `task build_wasm_hello`. The
rest of the plan below stands as written.

## What a wasm build buys

* **Static hosting.** `index.html`, `wasm_exec.js`, `app.wasm` and the frontend
  bundle are files. GitHub Pages, S3 or a CDN serve them; there is no server to
  run, scale or pay for.
* **A live book.** The demo could run inside the documentation site instead of
  being a screenshot.
* **Local data.** A file the user picks never leaves the tab, which the web
  executor cannot promise.

Non-goals: a wasm build does not replace `tgexec.WebExecutor`. Anything that
reads the filesystem, talks to a database, holds a secret or reaches a service
on the local network stays server-side. The app's Go code is also shipped to
the visitor, so it is public.

## What the repository already gives us

Most of the work is done, by the desktop executor.

1. **`tgframe.Session` is transport-agnostic.** It needs a `SendPackFunc` and
   raw event bytes, and nothing else. `tgexec` feeds it a websocket, `tgwails`
   feeds it bound methods and an event queue. A third transport is the same
   contract again.
2. **The frontend is transport-agnostic too.** `App` takes `update`, `upload`,
   `pageName` and `onNavigate` props; `@toolgui-web/lib` holds the components,
   and the transport lives in a thin app around it. `toolgui-web/wails` is
   about 150 lines of that glue.
3. **Both lanes already carry the same wire format** — the JSON the websocket
   sends is the JSON the Wails binding sends as a string.
4. **`SetHashPageNameMode` exists.** `#/{name}` routing is what a static host
   needs, and it is already wired through the side nav.
5. **The per-transport layout is a convention**: a Go package, a frontend app
   under `toolgui-web/`, a `Taskfile` target, and a stub in
   `scripts/stub-assets.sh` so `go build` works without a yarn build.

## Measurements

Go 1.27.1, `GOOS=js GOARCH=wasm`. "Minimal" is `tgframe` + `tgcomp` with a
textbox, a text and a button, plus a `syscall/js` bridge.

| Build | `.wasm` | gzip -9 |
| --- | --- | --- |
| Minimal | 5.31 MiB | 1.44 MiB |
| Minimal, `-ldflags="-s -w"` | 5.21 MiB | 1.41 MiB |
| Minimal + `tgexec.NewWebExecutor` | 7.50 MiB | 2.19 MiB |
| `cmd/toolgui-demo` (jpeg, zip, md5, embed) | 15.8 MiB | 4.07 MiB |

Reading them:

* **The floor is the Go runtime**, ~1.4 MiB gzipped, and the smallest app pays
  it in full. Stripping symbols saves ~2%; it is not the lever.
* **A wasm entry point must not construct a `WebExecutor`.** `toolguiweb`
  embeds the built frontend, currently 2.2 MiB; the linker drops it while
  nothing references it (a blank import of `tgexec` measures 5.29 MiB), and
  `NewWebExecutor` is what references it. That is the +2.2 MiB in the third
  row, and it is the JS bundle the page already loaded from disk. The demo
  carries it for the same reason.
* **What the app itself imports dominates the rest.** The demo is well past the
  floor on its own dependencies, not the framework's.

Startup, headless Chromium over localhost, wasm in a worker: fetch 79 ms,
`WebAssembly.instantiate` 17 ms, `go.run` plus the first page run 33 ms. So
roughly 50 ms of CPU on top of whatever the download costs — a 1.4 MiB gzipped
transfer, once, cacheable.

## What the probes settled

Prototypes under Node 22 and Chromium, driving a real `tgframe.Session` through
a `syscall/js` bridge:

* `GOOS=js GOARCH=wasm go build ./...` **passes on the tree as it stands**.
  `tgexec` compiles for wasm; its listener just cannot run there.
* A `Session` over the bridge produces the same pack stream as the websocket:
  `ready`, the creates, then the result.
* **The interrupt path works.** An event delivered while a page function slept
  cut the run with `ErrUpdateInterrupt` and started the next one, exactly as on
  the server.
* **A blocking `js.FuncOf` callback does not deadlock, but it returns early.**
  The Go side kept running after the callback had returned to JavaScript. So
  the bridge cannot mimic the Wails bindings, whose methods return promises the
  frontend awaits: it has to be fire-and-forget calls plus a pack callback,
  which is what the websocket transport already looks like.
* **Go wasm runs in a Web Worker**, with packs crossing as `postMessage`.

## Options

| Option | New frontend code | UI blocked by a slow page func | Size | Risk |
| --- | --- | --- | --- | --- |
| **A. Bridge on the main thread** | one small app (Wails-shaped) | yes | floor | low |
| **B. Bridge in a Web Worker** | same app + ~30 lines of worker glue | no | floor | low |
| **C. Shim `WebSocket`/`fetch`, reuse `toolgui-web/app`** | none, but a fake socket | yes | floor | medium |
| **D. TinyGo** | as A or B | as A or B | ~10x smaller in principle | high |

**A. Main-thread bridge.** Go registers functions on `globalThis`; the frontend
calls them and receives packs through a callback. Closest to `tgwails`. The
catch is that Go's wasm is single-threaded and shares the browser's main
thread: a page function that computes for a second freezes the tab, and the
interrupt only lands at a component send, so a tight loop with no sends cannot
be cut at all.

**B. Worker bridge — recommended.** The same Go side, loaded by a worker whose
glue turns `postMessage` into calls and pack callbacks into messages. The page
function then runs off the main thread, so React keeps rendering and the input
that interrupts the run is still deliverable. It also matches the shape the
frontend already expects from the websocket: asynchronous, no return values. Go
never touches the DOM, so nothing is lost by leaving the main thread.

**C. Transport shim.** Replacing `window.WebSocket` with a fake backed by wasm
would let `toolgui-web/app` run unchanged. Attractive until the details: the
app also health-checks `/api/health`, uploads to `/api/files` and reconnects
with backoff, so `fetch` needs shimming too, and the reconnect state machine
becomes dead weight pretending a socket dropped. It trades a small, honest app
for a fake of a protocol.

**D. TinyGo.** The only real answer to the 1.4 MiB floor, but `tgframe` leans on
`encoding/json` and `any` component structs, which is exactly where TinyGo's
reflection support is thinnest. Worth a spike once the transport exists, not
before.

## Sketch of the recommended shape

**Go — `toolgui/tgwasm`, behind `//go:build js && wasm`.** It needs only
`syscall/js`, so unlike `toolgui-wails` it does not need its own module.

```go
//go:build js && wasm

app := tgframe.NewApp()
app.AddPage("index", "Index", Main)
tgwasm.NewExecutor(app).Run()   // registers the bridge, then blocks
```

The bridge it installs on `globalThis`, mirroring what the Wails backend binds:

| JS | Go | Server counterpart |
| --- | --- | --- |
| `toolgui.appConf()` | marshal `app.AppConf()` | `GET /api/app` |
| `toolgui.start(page)` | new `State` + `Session` | the update socket opening |
| `toolgui.update(json)` | `session.HandleRawEvent` | a socket message |
| `toolgui.uploadFile(name, b64)` | `state.SetFile` | `POST /api/files` |
| `toolgui.onPack(cb)` | the `SendPackFunc` | a socket message back |

Every call returns immediately; results arrive as packs, because a blocking
callback returns to JavaScript early.

**Frontend — `toolgui-web/wasm`.** `WasmApp.tsx` and `api/backend.ts` in the
shape of the Wails app, plus `worker.ts` that loads `wasm_exec.js` and
`app.wasm` and relays both directions. Builds to `toolgui-web/wasm/build`.

**Distribution — `cmd/toolgui-wasm`.** A host-side CLI (`go tool`, as the wails
CLI is wired today) that runs `GOOS=js GOARCH=wasm go build` on the user's
package, copies `wasm_exec.js` out of the user's `$(go env GOROOT)/lib/wasm`
— never a vendored copy, it must match the toolchain that built the binary —
and writes the embedded frontend next to it:

```
dist/index.html  static/…  wasm_exec.js  app.wasm
```

Serving notes worth documenting: `application/wasm` so
`instantiateStreaming` works, gzip or brotli on the `.wasm`, a content hash in
its name, and `SetHashPageNameMode(true)` unless the host can rewrite unknown
paths to `index.html`.

## How a user builds it

Two entry points around one page function, split by build tag. Both halves
compile today — `go build ./...` picks the server one, `GOOS=js GOARCH=wasm`
picks the other:

```go
// page.go — no build tag, shared.
func newApp() *tgframe.App { ... }

// main_server.go
//go:build !(js && wasm)
func main() { tgexec.NewWebExecutor(newApp()).StartService(":3000") }

// main_wasm.go
//go:build js && wasm
func main() { tgwasm.NewExecutor(newApp()).Run() }
```

With phase 1 only, the rest is by hand:

```shell
GOOS=js GOARCH=wasm go build -o dist/app.wasm ./cmd/myapp
cp "$(go env GOROOT)/lib/wasm/wasm_exec.js" dist/   # must match the toolchain
cp -r "$TOOLGUI/toolgui-web/wasm/build/"* dist/     # index.html + static/
python3 -m http.server -d dist                      # or any static host
```

Phase 2 is that script, minus the parts a user should not have to know:

```shell
go get github.com/voilelab/toolgui
go tool toolgui-wasm build ./cmd/myapp -o dist
```

It runs the same `go build`, takes `wasm_exec.js` from the caller's `GOROOT`,
writes the frontend it embeds, and leaves a `dist/` a static host serves as is
— including from a Pages workflow, which is then three lines: set up Go, run
the CLI, upload `dist/`.

Calling `SetHashPageNameMode(true)` is the one thing the app itself has to do,
unless the host can rewrite unknown paths to `index.html`.

## What the browser takes away

Things the docs would have to state plainly:

* One session per tab, created on load. No `UUIDMap`, no reconnect, no resume:
  a reload starts from an empty state.
* No filesystem, no listening socket. `net/http` requests become `fetch`, so
  CORS applies to every call the page function makes.
* `GOMAXPROCS` is 1. Goroutines interleave, nothing runs in parallel.
* No timezone database unless the app imports `time/tzdata` (~450 KiB).
* Uploads cross as bytes into memory; `MaxUploadSize` is replaced by whatever
  the tab can hold.
* The app binary is public. No secrets, no private data sources.
* `DownloadButton` and `Iframe` are unaffected — both already resolve in the
  browser.

## Phasing

1. `tgwasm` + `toolgui-web/wasm` + a hello example, assembled by hand behind a
   `Taskfile` target. Worker from the start.
2. `cmd/toolgui-wasm` so a user gets a `dist/` without knowing any of it —
   which is what adds a `stub-assets.sh` entry, since the CLI embeds the
   frontend — plus a book page next to the desktop one and CI that builds the
   example.
3. Publish the demo to Pages beside the book; then look at size (TinyGo, or
   trimming what the demo imports).

## Open questions

* Does `tgwasm` belong in the main module (no new dependency, but a build-tagged
  package users cannot import on other platforms) or in a `toolgui-wasm`
  module, as the desktop executor does?
* Should `toolgui-web/app` and `toolgui-web/wasm` eventually be one app that
  picks a transport at runtime, rather than two entry points around the same
  library?
* Is a worker acceptable as the only mode, or does an escape hatch for the main
  thread earn its keep?
