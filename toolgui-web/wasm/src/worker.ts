// The worker the Go program runs in. Keeping it off the main thread is what
// lets the UI stay responsive while a page function runs: Go's wasm is
// single-threaded and never touches the DOM, so it has no reason to share the
// browser's thread.
//
// The main thread posts calls, packs come back, and a picked File crosses as
// itself: a structured clone of one hands over the blob it is backed by
// without copying its bytes.

// Bridge is the Go side, published on the worker's global object by the wasm
// program. See toolgui/tgwasm.
interface Bridge {
  appConf(): string
  onPack(callback: (packJSON: string) => void): void
  start(pageName: string): void
  update(eventJSON: string): void
  downloadFile(token: string): string
  newUpload(): string
  uploadFile(componentID: string, name: string, slot: string,
    handle: FileSystemSyncAccessHandle): string
  cancelUpload(slot: string): void
}

// UploadSlot is where Go reserved room for one upload, as newUpload answers
// it: a directory named from the origin private file system's root down, and
// the file to write inside it.
interface UploadSlot {
  dir?: string[]
  name?: string
  error?: string
}

// DownloadSlot is where the file behind a download token is, as downloadFile
// answers it. Same shape as an upload's: the two are the halves of one
// filesystem, read the same way.
type DownloadSlot = UploadSlot

// The synchronous side of the origin private file system, which the dom lib
// does not describe: it exists in a dedicated worker and nowhere else, and it
// is how the wasm program reads an upload back. Only what crosses to Go is
// named here -- Go does the reading through it.
interface FileSystemSyncAccessHandle {
  close(): void
}

// getFileHandle answers a handle the dom lib knows, minus the one method that
// only a dedicated worker has.
type SyncFileHandle = FileSystemFileHandle & {
  createSyncAccessHandle(): Promise<FileSystemSyncAccessHandle>
}

// The worker globals the dom lib does not describe.
interface WorkerCtx {
  importScripts(...urls: string[]): void
  postMessage(message: any): void
  onmessage: ((event: MessageEvent) => void) | null
  Go: new () => { importObject: WebAssembly.Imports, run(instance: WebAssembly.Instance): void }
  toolgui?: Bridge

  // The display mode, for a program that lays its page out differently in a
  // frame. Set before the wasm program runs, which is the only moment Go can
  // read it: the query string it comes from is the page's, and a worker's own
  // location is this script.
  toolguiEmbed?: boolean
}

const ctx = self as unknown as WorkerCtx

let bridge: Bridge | null = null

ctx.onmessage = (event: MessageEvent) => {
  const msg = event.data

  switch (msg.kind) {
    case 'init':
      ctx.toolguiEmbed = !!msg.embed
      boot(msg.wasmExecURL, msg.wasmURL).catch((e) => {
        ctx.postMessage({ kind: 'failed', error: String(e) })
      })
      break

    case 'call':
      call(msg.id, msg.fn, msg.args)
      break

    case 'upload':
      upload(msg.id, msg.componentID, msg.file)
      break

    case 'download':
      download(msg.id, msg.token)
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

// upload copies a picked file into the file system Go reads uploads from, and
// hands the finished file over.
//
// The copy is a stream. Nothing here ever holds the file: file.stream() reads
// it off disk a chunk at a time and the writable stream puts each chunk away,
// so a file of any size costs the tab a chunk rather than several copies of
// itself.
//
// The two halves cannot be swapped. A writable stream and a sync access handle
// are exclusive holds on the same file, and Go reads through the second, so the
// first has to be closed before Go is handed anything -- pipeTo closes it, and
// createSyncAccessHandle would answer NoModificationAllowedError if it had not.
// Writing from Go instead is no way around it either: a writable stream is what
// keeps the bytes out of the heap, and Go cannot await one.
async function upload(id: number, componentID: string, file: File) {
  if (!bridge) {
    ctx.postMessage({ kind: 'return', id, error: 'the wasm program is not running' })
    return
  }

  let slot: UploadSlot
  try {
    slot = JSON.parse(bridge.newUpload())
  } catch (e) {
    ctx.postMessage({ kind: 'return', id, error: String(e) })
    return
  }

  if (slot.error || !slot.name || !slot.dir) {
    ctx.postMessage({
      kind: 'return', id,
      error: slot.error || 'nowhere to store the upload',
    })
    return
  }

  try {
    const target = await slotFile(slot, true)

    // pipeTo closes the writable when the file ends, and aborts it on a
    // failure anywhere -- which leaves the file as it was rather than holding
    // half of one, because a writable stream writes to a swap file and only
    // puts it in place on close.
    await file.stream().pipeTo(await target.createWritable({ keepExistingData: false }))

    const handle = await target.createSyncAccessHandle()
    const error = bridge.uploadFile(componentID, file.name, slot.name, handle)

    ctx.postMessage({ kind: 'return', id, error: error || undefined })
  } catch (e) {
    // Out of quota, cancelled, or a stream that broke: Go drops the
    // reservation, so nothing half written is left behind or read. It also
    // closes a handle it would not take, so there is none to lose here.
    bridge.cancelUpload(slot.name)
    ctx.postMessage({ kind: 'return', id, error: String(e) })
  }
}

// download hands back the file a token names, as a File the main thread can
// make a blob URL of.
//
// Nothing is copied. getFile answers a blob backed by what is in the origin
// private file system, and a structured clone of one carries that backing
// rather than the bytes -- so a download of any size costs this worker and the
// message nothing, which is the whole point of not putting it in the pack.
//
// Go holds a sync access handle on the file. That is exclusive against a
// second handle and against a writable stream, not against this read, and Go
// wrote the file before it handed the token out.
async function download(id: number, token: string) {
  if (!bridge) {
    ctx.postMessage({ kind: 'return', id, error: 'the wasm program is not running' })
    return
  }

  let slot: DownloadSlot
  try {
    slot = JSON.parse(bridge.downloadFile(token))
  } catch (e) {
    ctx.postMessage({ kind: 'return', id, error: String(e) })
    return
  }

  if (slot.error || !slot.name || !slot.dir) {
    ctx.postMessage({
      kind: 'return', id,
      error: slot.error || 'nowhere to read the download from',
    })
    return
  }

  try {
    const handle = await slotFile(slot, false)
    ctx.postMessage({ kind: 'return', id, value: await handle.getFile() })
  } catch (e) {
    ctx.postMessage({ kind: 'return', id, error: String(e) })
  }
}

// slotFile walks to the file a slot names. An upload's is created here -- Go
// named it and deliberately never opened it, so the file is this side's to
// write until it is handed back -- while a download's is Go's own and has to
// be there already: creating one would answer an empty file where a token
// naming nothing belongs.
async function slotFile(slot: UploadSlot | DownloadSlot,
  create: boolean): Promise<SyncFileHandle> {
  let dir = await navigator.storage.getDirectory()
  for (const name of slot.dir) {
    dir = await dir.getDirectoryHandle(name)
  }

  return await dir.getFileHandle(slot.name, { create }) as SyncFileHandle
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
