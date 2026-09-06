import React, { useRef, useEffect, useCallback } from "react"

import { Props } from '../component_interface'

export function TIframe({ node, update, upload, theme }: Props) {
  const sandbox = node.props.script ? "allow-scripts allow-same-origin" : "allow-same-origin"

  const iframeRef = useRef<HTMLIFrameElement>(null)

  // Exposed to the guest as window.update(value). The id is fixed to this
  // iframe's own id, so the guest cannot forge events for other components.
  const sendUpdate = useCallback((value: any) => {
    update({ type: "iframe", id: node.props.id, value: value })
  }, [update, node.props.id])

  const inject = useCallback(() => {
    const contentWindow = iframeRef.current?.contentWindow
    if (!contentWindow) {
      return
    }

    contentWindow['props'] = node.props
    contentWindow['update'] = sendUpdate
    contentWindow['upload'] = upload
    contentWindow['theme'] = theme
  }, [node.props, sendUpdate, upload, theme])

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

  return (
    <iframe
      ref={iframeRef}
      id={node.props.id}
      name={node.props.id}
      sandbox={sandbox}
      srcDoc={node.props.html}
      onLoad={inject}
      style={{
        width: node.props.width,
        height: node.props.height,
        border: 'none',
      }}
    />
  )
}
