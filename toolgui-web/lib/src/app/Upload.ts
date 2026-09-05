// UploadResult is the transport-agnostic result of uploading a file.
// A transport (http, desktop binding, ...) maps its own response onto it.
export interface UploadResult {
  ok: boolean
  error?: string
}

export type UploadFunc = (file: File) => Promise<UploadResult>
