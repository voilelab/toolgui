import React, { useRef, useEffect, useCallback, useState } from "react"

import { UpdateEvent } from "../../app/UpdateEvent"
import { UploadFunc } from "../../app/Upload"
import { ThemeMode } from "../../util/theme"
import { GUEST_SCRIPT } from './iframe_guest'

export const PROTOCOL_VERSION = 1

export interface GuestFrameProps {
  // id is the component's own id. It is what a value the guest sends is keyed
  // by, so a guest without one cannot send anything back.
  id: string

  // guestProps is handed to the guest on every render message.
  guestProps: any

  theme: ThemeMode

  // html is the guest document. The helper is prepended to it here, so the
  // guest gets its bridge without any cross-origin access.
  html: string

  // script allows the guest to run javascript. Without it there is no bridge
  // and the guest is a plain document.
  script: boolean

  // width and height are css. A height of "auto" tracks the height the guest
  // reports through window.toolgui.autoHeight().
  width: string
  height: string

  update: (event: UpdateEvent) => void
  upload: UploadFunc
}

// GuestFrame is the host half of the window.toolgui bridge: a sandboxed frame
// running a guest document, and the postMessage conversation with it. Iframe
// and Plugin are the two components built on it, and differ only in what the
// guest document is.
export function GuestFrame({
  id, guestProps, theme, html, script, width, height, update, upload,
}: GuestFrameProps) {
  // No allow-same-origin: the guest gets an opaque origin, so it cannot reach
  // the app's DOM, cookies or storage, and cannot shed its own sandbox.
  // Everything it needs goes over postMessage instead.
  const sandbox = script ? "allow-scripts" : ""

  const iframeRef = useRef<HTMLIFrameElement>(null)

  // Set once the guest has announced itself. postMessage does not queue, so
  // anything sent before that is dropped.
  const guestReady = useRef(false)

  // Height reported by the guest, used when the component asks for "auto".
  const [contentHeight, setContentHeight] = useState<number | null>(null)

  const post = useCallback((message: any) => {
    const contentWindow = iframeRef.current?.contentWindow
    if (!contentWindow) {
      return
    }

    // An opaque-origin guest has no origin to name, so "*" is the only option.
    // Nothing sensitive goes this way: props and theme are the guest's own.
    contentWindow.postMessage({ ...message, toolgui: PROTOCOL_VERSION }, '*')
  }, [])

  const postRender = useCallback(() => {
    post({ type: 'render', id: id, props: guestProps, theme: theme })
  }, [post, id, guestProps, theme])

  const sendUpload = useCallback(async (requestID: string, file: any) => {
    if (!(file instanceof File)) {
      post({ type: 'upload_result', requestID: requestID, ok: false, error: 'not a file' })
      return
    }

    const result = await upload(file)
    post({
      type: 'upload_result',
      requestID: requestID,
      ok: result.ok,
      error: result.error,
    })
  }, [post, upload])

  useEffect(() => {
    const onMessage = (event: MessageEvent) => {
      // event.origin is "null" for every opaque-origin frame, so it cannot
      // tell one sender from another. The window reference can.
      if (event.source !== iframeRef.current?.contentWindow) {
        return
      }

      const data = event.data
      if (!data || data.toolgui !== PROTOCOL_VERSION) {
        return
      }

      switch (data.type) {
        case 'ready':
          guestReady.current = true
          postRender()
          break

        case 'update':
          // The id is ours, never the guest's: a guest can only write to its
          // own state. Without one it has no state to write to.
          if (id) {
            update({ type: "custom", id: id, value: data.value })
          }
          break

        case 'resize':
          if (typeof data.height === 'number') {
            setContentHeight(data.height)
          }
          break

        case 'upload':
          sendUpload(data.requestID, data.file)
          break
      }
    }

    window.addEventListener('message', onMessage)
    return () => { window.removeEventListener('message', onMessage) }
  }, [postRender, update, id, sendUpload])

  // Push the new values to a guest that is already listening. A guest that is
  // still loading gets them from the render that answers its ready.
  useEffect(() => {
    if (guestReady.current) {
      postRender()
    }
  }, [postRender])

  // Without allow-scripts the helper cannot run, so there is no point shipping
  // it. The host builds srcDoc, so prepending needs no access to the guest.
  const srcDoc = script
    ? '<script>' + GUEST_SCRIPT + '</script>' + html
    : html

  const frameHeight = height === 'auto'
    ? (contentHeight === null ? undefined : contentHeight)
    : height

  return (
    <iframe
      ref={iframeRef}
      id={id}
      sandbox={sandbox}
      srcDoc={srcDoc}
      style={{
        width: width,
        height: frameHeight,
        border: 'none',
      }}
    />
  )
}
