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

## What crosses the boundary

The Go program runs in a Web Worker and publishes `globalThis.toolgui`:

| Web | Browser |
| --- | --- |
| `GET /api/app` | `toolgui.appConf()` |
| update websocket | `toolgui.update(eventJSON)` + the `toolgui.onPack` callback |
| `POST /api/files` | `toolgui.uploadFile(name, base64)` |
| a page load | `toolgui.start(pageName)` |

Payloads cross as JSON strings — the same ones the websocket carries, so both
transports share a wire format.

Every call returns at once. A call into Go that blocks hands control back to
JavaScript before its work is done, so results come back as packs instead of
return values; only `uploadFile`, which cannot block, answers with an error
string.

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
