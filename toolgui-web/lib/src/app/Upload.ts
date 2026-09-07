// UploadResult is the transport-agnostic result of uploading a file.
// A transport (http, desktop binding, ...) maps its own response onto it.
export interface UploadResult {
  ok: boolean
  error?: string
}

// componentID says which fileupload the file belongs to, so a transport can
// store it per component instead of per file name.
export type UploadFunc = (file: File, componentID: string) => Promise<UploadResult>
