import { AppConf, UpdateEvent, UploadResult } from "@toolgui-web/lib"

// Backend is the Go struct Wails binds. Every bound method returns a promise.
// Payloads cross as JSON strings, the same ones the websocket transport
// carries, so both lanes share a wire format.
interface Backend {
  AppConf(): Promise<string>
  Start(pageName: string): Promise<void>
  Update(eventJSON: string): Promise<void>
  UploadFileStart(name: string): Promise<string>
  UploadFileChunk(uploadID: string, dataBase64: string): Promise<void>
  UploadFileFinish(componentID: string, uploadID: string): Promise<void>
}

// WailsRuntime is the slice of window.runtime this adapter uses.
interface WailsRuntime {
  EventsOn(eventName: string, callback: (...data: any[]) => void): () => void
}

declare global {
  interface Window {
    // Wails names bindings after the Go package and struct they came from.
    go: { tgwails: { ToolGUI: Backend } }
    runtime: WailsRuntime
  }
}

export function backend(): Backend {
  return window.go.tgwails.ToolGUI
}

export function onEvent(eventName: string, callback: (data: string) => void) {
  window.runtime.EventsOn(eventName, callback)
}

// getAppConf is the desktop counterpart of GET /api/app.
export async function getAppConf(): Promise<AppConf> {
  return JSON.parse(await backend().AppConf())
}

export function sendEvent(event: UpdateEvent): Promise<void> {
  return backend().Update(JSON.stringify(event))
}

// CHUNK_SIZE is how much of a file crosses the bridge at a time. Bindings
// take strings, so each chunk costs its own size again as base64.
const CHUNK_SIZE = 4 * 1024 * 1024

// uploadFile is the desktop counterpart of POST /api/files. It sends the file
// in chunks: reading it whole would hold it as a blob, as base64 and as bytes
// on the Go side all at once.
export async function uploadFile(file: File, componentID: string): Promise<UploadResult> {
  try {
    // The component gets the file only once every chunk has landed, so a
    // second pick while this one is still going replaces it rather than
    // being spliced into it.
    const uploadID = await backend().UploadFileStart(file.name)

    for (let offset = 0; offset < file.size; offset += CHUNK_SIZE) {
      const chunk = file.slice(offset, offset + CHUNK_SIZE)
      await backend().UploadFileChunk(uploadID, await toBase64(chunk))
    }

    await backend().UploadFileFinish(componentID, uploadID)
    return { ok: true }
  } catch (e) {
    return { ok: false, error: String(e) }
  }
}

// toBase64 drops the "data:<type>;base64," prefix FileReader adds.
function toBase64(file: Blob): Promise<string> {
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
