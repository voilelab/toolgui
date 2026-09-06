import React from "react"

import { Props } from '../component_interface'

export function TLink({ node }: Props) {
  return (
    <a id={node.props.id || undefined} href={node.props.url}>
      {node.props.text}
    </a>
  )
}
