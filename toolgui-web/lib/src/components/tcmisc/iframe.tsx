import React, { useMemo } from "react"

import { Props } from '../component_interface'
import { GuestFrame } from './guest_frame'

export function TIframe({ node, update, upload, theme }: Props) {
  // Send the props without the html: the guest is that html, and shipping it
  // back is a second copy of the document. Memoised, or every render of this
  // component would look like new props and re-post them to the guest.
  const guestProps = useMemo(() => {
    const { html, ...rest } = node.props
    return rest
  }, [node.props])

  return (
    <GuestFrame
      id={node.props.id}
      guestProps={guestProps}
      theme={theme}
      html={node.props.html || ''}
      script={node.props.script}
      width={node.props.width}
      height={node.props.height}
      update={update}
      upload={upload} />
  )
}
