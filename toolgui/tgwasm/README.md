# tgwasm

Runs a ToolGUI app in the browser, compiled to WebAssembly, instead of serving
it over HTTP.

```go
//go:build js && wasm

app := tgframe.NewApp()
app.AddPage("index", "Index", Main)
app.SetHashPageNameMode(true)

tgwasm.NewExecutor(app).Run()
```

The pages, components and state are the same as the web executor's: only the
transport changes. `tgframe.Session` does the work either way, and the frontend
is the same `@toolgui-web/lib` components with a different adapter around them.

The package is behind `//go:build js && wasm`, so it costs nothing on any other
platform and needs no dependency beyond `syscall/js` — which is why it lives in
the main module, unlike `toolgui-wails`.

## The Go program has to run in a dedicated Web Worker

Not a preference — a requirement, and the host page is responsible for it.
`@toolgui-web/wasm` boots the binary in `worker.ts`, and a page that hosts the
binary itself has to do the same.

Two reasons, and only the first is a hard one:

* Uploads are kept in the origin private file system, and the only way to read
  and write one without awaiting a promise is
  `FileSystemFileHandle.createSyncAccessHandle`, which exists in a dedicated
  worker and nowhere else. The store cannot await anything — `uploadFile`
  arrives on the JavaScript callback stack, where a Go function that blocks
  holds the event loop the promise is waiting on — so on the page's thread
  every upload fails. `tgframe` says as much rather than guessing: *no
  synchronous file access here: the Go program has to run in a dedicated Web
  Worker*.
* Off the page's thread, a page function that takes a while leaves the UI
  responsive. Go's wasm is single-threaded and never touches the DOM, so it has
  no reason to share the browser's thread anyway.

## What crosses the boundary

The worker publishes `globalThis.toolgui`:

| Web | Browser |
| --- | --- |
| `GET /api/app` | `toolgui.appConf()` |
| update websocket | `toolgui.update(eventJSON)` + the `toolgui.onPack` callback |
| `POST /api/files` | `toolgui.uploadFile(componentID, name, base64)` |
| a page load | `toolgui.start(pageName)` |

Payloads cross as JSON strings — the same ones the websocket carries, so both
transports share a wire format.

Every call returns at once. A call into Go that blocks hands control back to
JavaScript before its work is done, so results come back as packs instead of
return values; only `uploadFile`, which cannot block, answers with an error
string.

Because it cannot block, the file store keeps a few files created and open
ahead of demand, and an upload takes one of those rather than waiting for the
promises that make one. There are always some by the time a user can have
picked a file: the pool fills while the page is still being drawn.

Once it has filled, taking a file never waits again — it fails instead. A
callback that took more than the pool holds would be waiting on promises that
cannot settle until it returns, so the store gives up with an error rather than
stopping the tab. One `uploadFile` takes one file and the browser sends them one
per task, so nothing on the transport comes near the limit; a callback of your
own that stores a stack of files at once would.

## Building

`cmd/toolgui-wasm` compiles the app and writes the site around it -- the
frontend it carries embedded, and the `wasm_exec.js` of the toolchain that
just compiled the binary:

```shell
go get -tool github.com/voilelab/toolgui/cmd/toolgui-wasm

go tool toolgui-wasm build -o dist ./cmd/myapp
go tool toolgui-wasm serve ./cmd/myapp
```

In this repository:

```shell
task build_wasm_hello   # the example, into example/hello/build
task run_wasm_hello     # the same, served at http://localhost:3000
task run_wasm_demo      # the component demo, in the browser
```

See [the book](https://voilelab.github.io/toolgui/hello-world/wasm.html) for
what a build produces, what the browser takes away and how to host it.

## Testing

The Go tests of this build run in a browser, because that is where the file
system they use is:

```shell
task test_wasm
```

It needs a Chrome or Chromium on `PATH`, or `TOOLGUI_BROWSER` pointing at one.
`scripts/wasmtest` is what drives it: a `go test -exec` wrapper that boots the
test binary in a dedicated worker and relays its output back. It runs `tgframe`
and not the rest of `./toolgui/...`, because `tgexec` serves over HTTP and a tab
has no socket to listen on.
