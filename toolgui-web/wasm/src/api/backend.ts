import { AppConf, DownloadResult, UpdateEvent, UploadResult } from "@toolgui-web/lib"

// Backend drives the wasm program in its worker. Calls go out as messages and
// come back by id; packs arrive on their own, the way the websocket transport
// delivers them.
export class Backend {
  private worker: Worker
  private pending = new Map<number, (msg: any) => void>()
  private nextID = 1
  private ready: Promise<void>

  // onPack is called for every pack, in the order the page produced them.
  // embed is the display mode, which the program is told at boot: a worker's
  // own location is this script, so it cannot read the page's query string.
  constructor(onPack: (pack: any) => void, embed: boolean = false) {
    this.worker = new Worker(new URL('../worker.ts', import.meta.url))

    let started: () => void
    let failed: (e: any) => void
    this.ready = new Promise((resolve, reject) => { started = resolve; failed = reject })

    this.worker.onmessage = (event: MessageEvent) => {
      const msg = event.data

      switch (msg.kind) {
        case 'ready':
          started()
          break

        case 'failed':
          failed(new Error(msg.error))
          break

        case 'pack':
          onPack(msg.pack)
          break

        case 'return': {
          const resolve = this.pending.get(msg.id)
          this.pending.delete(msg.id)
          resolve?.(msg)
          break
        }

        default:
          console.error('unknown message', msg.kind)
      }
    }

    // Both files sit next to index.html, wherever the app is hosted.
    this.worker.postMessage({
      kind: 'init',
      wasmExecURL: assetURL('wasm_exec.js'),
      wasmURL: assetURL('app.wasm'),
      embed,
    })
  }

  // send posts one message the worker answers by id, whatever it is.
  private async send(msg: any): Promise<any> {
    await this.ready

    const id = this.nextID++
    return new Promise((resolve, reject) => {
      this.pending.set(id, (answer) => {
        if (answer.error) {
          reject(new Error(answer.error))
          return
        }

        resolve(answer.value)
      })

      this.worker.postMessage({ ...msg, id })
    })
  }

  private call(fn: string, ...args: any[]): Promise<any> {
    return this.send({ kind: 'call', fn, args })
  }

  // appConf is the browser counterpart of GET /api/app.
  async appConf(): Promise<AppConf> {
    return JSON.parse(await this.call('appConf'))
  }

  // start opens a session on a page and draws it once.
  start(pageName: string): Promise<void> {
    return this.call('start', pageName)
  }

  update(event: UpdateEvent): Promise<void> {
    return this.call('update', JSON.stringify(event))
  }

  // uploadFile is the browser counterpart of POST /api/files.
  //
  // The File itself crosses, not its content. A structured clone of one hands
  // the worker the same blob this thread holds -- the browser has it on disk
  // already -- so nothing is copied here and nothing this size is ever a
  // string. The worker streams it into the file system the wasm program reads
  // uploads from; see worker.ts.
  async uploadFile(file: File, componentID: string): Promise<UploadResult> {
    try {
      await this.send({ kind: 'upload', componentID, file })
      return { ok: true }
    } catch (e) {
      return { ok: false, error: String(e) }
    }
  }

  // downloadFile is the browser counterpart of GET /api/files.
  //
  // Nothing crosses but the token and, coming back, the file the worker read
  // out of the origin private file system: a structured clone of it carries
  // the blob's backing rather than its bytes, so a download costs this thread
  // nothing whatever it weighs.
  async downloadFile(token: string): Promise<DownloadResult> {
    try {
      const file = await this.send({ kind: 'download', token })
      return { ok: true, blob: file }
    } catch (e) {
      return { ok: false, error: String(e) }
    }
  }
}

function assetURL(name: string): string {
  return new URL(name, document.baseURI).href
}
