import React from "react"

import { Props } from "../component_interface"

export function TText({ node }: Props) {
  return (
    <div id={node.props.id || undefined}>{node.props.text}</div>
  )
}
