// DownloadResult is the transport-agnostic result of fetching a file a run
// offered. The blob is backed by whatever the transport read it out of -- a
// response body, a file in the origin private file system -- so the bytes are
// the browser's to keep wherever it keeps them, not a string in the heap.
export interface DownloadResult {
  ok: boolean
  blob?: Blob
  error?: string
}

// token names one file the page offered. It is unguessable and belongs to the
// state that offered it, so a transport sends it with whatever already
// identifies the connection rather than on its own.
export type DownloadFunc = (token: string) => Promise<DownloadResult>
