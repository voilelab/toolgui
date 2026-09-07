// The worker the Go program runs in. Keeping it off the main thread is what
// lets the UI stay responsive while a page function runs: Go's wasm is
// single-threaded and never touches the DOM, so it has no reason to share the
// browser's thread.
//
// The main thread posts calls and receives packs; nothing else crosses.

// Bridge is the Go side, published on the worker's global object by the wasm
// program. See toolgui/tgwasm.
interface Bridge {
  appConf(): string
  onPack(callback: (packJSON: string) => void): void
  start(pageName: string): void
  update(eventJSON: string): void
  uploadFile(componentID: string, name: string, dataBase64: string): string
}

// The worker globals the dom lib does not describe.
interface WorkerCtx {
  importScripts(...urls: string[]): void
  postMessage(message: any): void
  onmessage: ((event: MessageEvent) => void) | null
  Go: new () => { importObject: WebAssembly.Imports, run(instance: WebAssembly.Instance): void }
  toolgui?: Bridge
}

const ctx = self as unknown as WorkerCtx

let bridge: Bridge | null = null

ctx.onmessage = (event: MessageEvent) => {
  const msg = event.data

  switch (msg.kind) {
    case 'init':
      boot(msg.wasmExecURL, msg.wasmURL).catch((e) => {
        ctx.postMessage({ kind: 'failed', error: String(e) })
      })
      break

    case 'call':
      call(msg.id, msg.fn, msg.args)
      break

    default:
      console.error('unknown message', msg.kind)
  }
}

async function boot(wasmExecURL: string, wasmURL: string) {
  // wasm_exec.js is the runtime shim of the toolchain that built app.wasm, so
  // it is fetched at runtime rather than bundled into this worker.
  ctx.importScripts(wasmExecURL)

  const go = new ctx.Go()
  const { instance } = await instantiate(wasmURL, go.importObject)

  // The Go program installs the bridge and then blocks, which is what hands
  // control back here.
  go.run(instance)

  bridge = ctx.toolgui
  if (!bridge) {
    throw new Error('the wasm program installed no bridge')
  }

  // Registered before the first start, or that run's packs are lost.
  bridge.onPack((packJSON: string) => {
    ctx.postMessage({ kind: 'pack', pack: JSON.parse(packJSON) })
  })

  ctx.postMessage({ kind: 'ready' })
}

async function instantiate(wasmURL: string, importObject: WebAssembly.Imports) {
  try {
    return await WebAssembly.instantiateStreaming(fetch(wasmURL), importObject)
  } catch {
    // A host that serves the binary as something other than application/wasm.
    const bs = await (await fetch(wasmURL)).arrayBuffer()
    return await WebAssembly.instantiate(bs, importObject)
  }
}

function call(id: number, fn: keyof Bridge, args: any[]) {
  if (!bridge) {
    ctx.postMessage({ kind: 'return', id, error: 'the wasm program is not running' })
    return
  }

  try {
    const value = (bridge[fn] as (...a: any[]) => any)(...args)
    ctx.postMessage({ kind: 'return', id, value })
  } catch (e) {
    ctx.postMessage({ kind: 'return', id, error: String(e) })
  }
}
