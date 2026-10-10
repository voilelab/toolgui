import React, { useCallback, useState } from "react"

import { SessionApp } from "@toolgui-web/lib"
import { Backend, LoadProgress } from "./api/backend"

// embedFromSearch reads the embed flag off the URL. embed is a display mode of
// the front end rather than something the app declares, so it lives in the
// address: the same binary serves a site and an iframe.
function embedFromSearch(search: string): boolean {
  const value = new URLSearchParams(search).get('embed')

  // Bare `?embed` counts. The two spellings that do not are the ones a caller
  // building the URL from a boolean would produce for "no".
  return value !== null && value !== '0' && value !== 'false'
}

// Read once: the query string cannot change without a page load, while the
// hash changes on every page.
const embed = embedFromSearch(window.location.search)

// WasmApp keeps the page in the URL hash: a static host cannot route paths.
export function WasmApp() {
  const [progress, setProgress] = useState<LoadProgress | null>(null)

  const connect = useCallback((onPack: (pack: any) => void) =>
    new Backend(onPack, embed, setProgress, window.location.search), [])

  return <SessionApp connect={connect} useHash embed={embed}
    loading={<Loading progress={progress} />} />
}

// Loading shows the download of app.wasm, which is most of a first visit.
function Loading({ progress }: { progress: LoadProgress | null }) {
  if (!progress) {
    return <div className="tg-loading"><progress /></div>
  }

  if (progress.done) {
    return <div className="tg-loading"><progress /><p>Starting…</p></div>
  }

  const { loaded, total } = progress
  return (
    <div className="tg-loading">
      {total > 0 ? <progress value={loaded} max={total} /> : <progress />}
      <p>Loading {formatMB(loaded)}{total > 0 ? ` / ${formatMB(total)}` : ''}</p>
    </div>
  )
}

function formatMB(n: number): string {
  return `${(n / (1024 * 1024)).toFixed(1)} MB`
}
