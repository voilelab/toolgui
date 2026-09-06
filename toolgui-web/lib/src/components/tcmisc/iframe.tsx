import React, { useRef, useEffect, useCallback, useState } from "react"

import { Props } from '../component_interface'
import { GUEST_SCRIPT } from './iframe_guest'

const PROTOCOL_VERSION = 1

export function TIframe({ node, update, upload, theme }: Props) {
  const sandbox = node.props.script ? "allow-scripts allow-same-origin" : "allow-same-origin"

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
    // Send the props without the html: the guest is that html, and shipping it
    // back is a second copy of the document.
    const { html, ...props } = node.props

    post({ type: 'render', id: node.props.id, props: props, theme: theme })
  }, [post, node.props, theme])

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
          // The id is ours, never the guest's: an iframe can only write to
          // its own state.
          update({ type: "iframe", id: node.props.id, value: data.value })
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
  }, [postRender, update, node.props.id, sendUpload])

  // Push the new values to a guest that is already listening. A guest that is
  // still loading gets them from the render that answers its ready.
  useEffect(() => {
    if (guestReady.current) {
      postRender()
    }
  }, [postRender])

  // Deprecated: the injected globals predate the postMessage transport and go
  // away with allow-same-origin. Prefer window.toolgui.
  const inject = useCallback(() => {
    const contentWindow = iframeRef.current?.contentWindow
    if (!contentWindow) {
      return
    }

    contentWindow['props'] = node.props
    contentWindow['update'] = (value: any) => {
      update({ type: "iframe", id: node.props.id, value: value })
    }
    contentWindow['upload'] = upload
    contentWindow['theme'] = theme
  }, [node.props, update, upload, theme])

  useEffect(() => {
    // srcDoc loads asynchronously, so right after a html change contentWindow
    // still points at the outgoing document. onLoad covers that case; this
    // effect only refreshes the values when they change without a reload.
    inject()

    return () => {
      // Re-read instead of capturing: the document may have reloaded since
      // the injection, and a detached iframe has no contentWindow at all.
      const contentWindow = iframeRef.current?.contentWindow
      if (!contentWindow) {
        return
      }

      contentWindow['props'] = null
      contentWindow['update'] = null
      contentWindow['upload'] = null
      contentWindow['theme'] = null
    }
  }, [inject])

  // Without allow-scripts the helper cannot run, so there is no point shipping
  // it. The host builds srcDoc, so prepending needs no access to the guest.
  const html = node.props.html || ''
  const srcDoc = node.props.script
    ? '<script>' + GUEST_SCRIPT + '</script>' + html
    : html

  const height = node.props.height === 'auto'
    ? (contentHeight === null ? undefined : contentHeight)
    : node.props.height

  return (
    <iframe
      ref={iframeRef}
      id={node.props.id}
      name={node.props.id}
      sandbox={sandbox}
      srcDoc={srcDoc}
      onLoad={inject}
      style={{
        width: node.props.width,
        height: height,
        border: 'none',
      }}
    />
  )
}
