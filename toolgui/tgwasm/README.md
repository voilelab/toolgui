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

* Uploads are kept in the origin private file system, and read back through
  `FileSystemFileHandle.createSyncAccessHandle`. That file system is a secure
  context's, so the site has to be served over `https` or from `localhost`; on
  plain `http` anywhere else there is nowhere to store an upload and the store
  says so rather than quietly falling back to the heap. A sync access handle is
  the only way to read and write one without awaiting a promise, and it exists
  in a dedicated worker and nowhere else. The store cannot await anything —
  `uploadFile` arrives on the JavaScript callback stack, where a Go function
  that blocks holds the event loop the promise is waiting on — so on the
  page's thread every upload fails. `tgframe` says as much rather than
  guessing: *no synchronous file access here: the Go program has to run in a
  dedicated Web Worker*.
* Off the page's thread, a page function that takes a while leaves the UI
  responsive. Go's wasm is single-threaded and never touches the DOM, so it has
  no reason to share the browser's thread anyway.

## What crosses the boundary

The worker publishes `globalThis.toolgui`:

| Web | Browser |
| --- | --- |
| `GET /api/app` | `toolgui.appConf()` |
| update websocket | `toolgui.update(eventJSON)` + the `toolgui.onPack` callback |
| `POST /api/files` | `toolgui.newUpload()`, then `toolgui.uploadFile(componentID, name, slot, handle)` or `toolgui.cancelUpload(slot)` |
| `GET /api/files` | `toolgui.downloadFile(token)` |
| a page load | `toolgui.start(pageName)` |

Payloads cross as JSON strings — the same ones the websocket carries, so both
transports share a wire format.

Every call returns at once. A call into Go that blocks hands control back to
JavaScript before its work is done, so results come back as packs instead of
return values; only the upload calls, which cannot block, answer with a value.

Because it cannot block, the file store keeps a few files created and open
ahead of demand, and a file the Go side writes takes one of those rather than
waiting for the promises that make one. There are always some by the time a
user can have picked a file: the pool fills while the page is still being
drawn.

Once it has filled, taking a file never waits again — it fails instead. A
callback that took more than the pool holds would be waiting on promises that
cannot settle until it returns, so the store gives up with an error rather than
stopping the tab. Nothing on the transport comes near that limit; a callback of
your own that stores a stack of files at once would.

The page talks back once, before any of that. The frontend's `?embed` — its
display mode, not something the app declares — is posted to the worker at
boot and set on `globalThis.toolguiEmbed`, and `tgwasm.Embedded()` reads it.
It is the one thing an app can ask about how it is being shown, and it is for
a page laid out differently in a frame. A worker's own location is its script
rather than the page's URL, which is why the flag is handed over instead of
read.

## An upload is written by the page and read by Go

No bytes cross the boundary. `newUpload` answers with a directory and a file
name, the page streams the picked file into that file with
`file.stream().pipeTo(writable)`, and `uploadFile` hands over a component ID,
the name the user's file had, and a sync access handle already open on what was
written. A page that gave up calls `cancelUpload` instead, and nothing half
written is left in the file system or reaches a component.

The split is the browser's doing, not a preference:

* A writable stream is the only way to copy a picked file without the whole of
  it passing through the tab's heap, and it is the page's to open. This used to
  be one `readAsDataURL` string handed to Go to decode — several copies of the
  file at once, two of them a third larger again — which is what stopped a
  large upload dead.
* A sync access handle is the only way to read a file back without awaiting a
  promise, which is what the store needs and cannot have on the callback stack.

The two are exclusive holds on the same file, so neither side can do both, and
the order is fixed: the page closes its writable stream, opens the handle, and
only then calls `uploadFile`. Asking for the handle while the write is still
open is `NoModificationAllowedError`. `TestBrowserUploadHoldsAreExclusive` pins
it.

## A download is written by Go and read by the page

The other direction is the easier half. `DownloadFile` writes the bytes into
the file store, which in this build is the origin private file system, and puts
only a token in the component. `downloadFile` turns that token back into a
directory and a file name, the worker opens the file with `getFile`, and the
main thread saves the blob that comes back through a URL of its own. No bytes
cross the boundary here either, and nothing of the file is ever a string in the
tab.

Go keeps its sync access handle on the file open throughout. That is exclusive
against a second handle and against a writable stream, but not against
`getFile`, and the file is written before its token reaches the page --
`TestBrowserDownloadIsReadableByThePage` pins it.

A token is looked up in the state that offered it and nowhere else, so it stops
working when a page switch replaces that state, or when a later run offers
different bytes.

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
