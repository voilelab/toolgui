import React from "react"

import { Props } from "../component_interface"

export function TText({ node }: Props) {
  return (
    <div>{node.props.text}</div>
  )
}
