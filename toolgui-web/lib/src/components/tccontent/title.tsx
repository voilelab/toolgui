import React from "react"

import { Props } from '../component_interface'

export function TTitle({ node }: Props) {
  return (
    <h1 id={node.props.id || undefined} className="title">
      {node.props.text}
    </h1>
  )
}
