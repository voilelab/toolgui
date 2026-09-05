import { AppConf, UpdateEvent, UploadResult } from "@toolgui-web/lib"

// Backend is the Go struct Wails binds. Every bound method returns a promise.
// Payloads cross as JSON strings, the same ones the websocket transport
// carries, so both lanes share a wire format.
interface Backend {
  AppConf(): Promise<string>
  Start(pageName: string): Promise<void>
  Update(eventJSON: string): Promise<void>
  UploadFile(name: string, dataBase64: string): Promise<void>
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

// uploadFile is the desktop counterpart of POST /api/files.
export async function uploadFile(file: File): Promise<UploadResult> {
  try {
    await backend().UploadFile(file.name, await toBase64(file))
    return { ok: true }
  } catch (e) {
    return { ok: false, error: String(e) }
  }
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
