import { AppConf, UpdateEvent, UploadResult } from "@toolgui-web/lib"

// Backend drives the wasm program in its worker. Calls go out as messages and
// come back by id; packs arrive on their own, the way the websocket transport
// delivers them.
export class Backend {
  private worker: Worker
  private pending = new Map<number, (msg: any) => void>()
  private nextID = 1
  private ready: Promise<void>

  // onPack is called for every pack, in the order the page produced them.
  constructor(onPack: (pack: any) => void) {
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
    })
  }

  private async call(fn: string, ...args: any[]): Promise<any> {
    await this.ready

    const id = this.nextID++
    return new Promise((resolve, reject) => {
      this.pending.set(id, (msg) => {
        if (msg.error) {
          reject(new Error(msg.error))
          return
        }

        resolve(msg.value)
      })

      this.worker.postMessage({ kind: 'call', id, fn, args })
    })
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
  async uploadFile(file: File, componentID: string): Promise<UploadResult> {
    try {
      const error = await this.call('uploadFile', componentID, file.name,
        await toBase64(file))
      return error ? { ok: false, error } : { ok: true }
    } catch (e) {
      return { ok: false, error: String(e) }
    }
  }
}

function assetURL(name: string): string {
  return new URL(name, document.baseURI).href
}

// toBase64 drops the "data:<type>;base64," prefix FileReader adds.
function toBase64(file: File): Promise<string> {
  return new Promise((resolve, reject) => {
    const reader = new FileReader()
    reader.onload = () => {
      const result = reader.result as string
      resolve(result.slice(result.indexOf(',') + 1))
    }
    reader.onerror = () => reject(reader.error)
    reader.readAsDataURL(file)
  })
}
