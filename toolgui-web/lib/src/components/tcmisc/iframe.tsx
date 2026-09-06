import React, { useRef, useEffect, useCallback } from "react"

import { Props } from '../component_interface'

export function TIframe({ node, update, upload, theme }: Props) {
  const sandbox = node.props.script ? "allow-scripts allow-same-origin" : "allow-same-origin"

  const iframeRef = useRef<HTMLIFrameElement>(null)


  useEffect(() => {
    const contentWindow = iframeRef.current?.contentWindow
    if (contentWindow) {
      contentWindow['props'] = node.props
      contentWindow['update'] = update
      contentWindow['upload'] = upload
      contentWindow['theme'] = theme
    }

    return () => {
      contentWindow['props'] = null
      contentWindow['update'] = null
      contentWindow['upload'] = null
      contentWindow['theme'] = null
    }
  }, [iframeRef, node.props, update, upload, theme])

  return (
    <iframe
      ref={iframeRef}
      id={node.props.id}
      name={node.props.id}
      sandbox={sandbox}
      srcDoc={node.props.html}
    />
  )
}
