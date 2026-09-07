import React, { useMemo } from "react"

import { Props } from '../component_interface'
import { GuestFrame } from './guest_frame'

// escapeAttr keeps a url from closing the attribute it sits in. The urls come
// from the page function, but the guest document is built as a string.
function escapeAttr(value: string) {
  return value
    .replace(/&/g, '&amp;')
    .replace(/"/g, '&quot;')
    .replace(/</g, '&lt;')
}

export function TPlugin({ node, update, upload, theme }: Props) {
  // The plugin is a script the app serves, so the guest document is only the
  // tags that load it. A classic script, not a module: modules are fetched
  // with cors, which an opaque-origin frame has no origin to pass. Deferred,
  // or it would run while the document is still head-only and every plugin
  // would have to wait for a body of its own.
  const html = useMemo(() => {
    const style = node.props.style
      ? `<link rel="stylesheet" href="${escapeAttr(node.props.style)}">`
      : ''

    const script = node.props.src
      ? `<script defer src="${escapeAttr(node.props.src)}"></script>`
      : ''

    return style + script
  }, [node.props.style, node.props.src])

  return (
    <GuestFrame
      id={node.props.id}
      guestProps={node.props.props}
      theme={theme}
      html={html}
      script={true}
      width={node.props.width}
      height={node.props.height}
      update={update}
      upload={upload} />
  )
}
