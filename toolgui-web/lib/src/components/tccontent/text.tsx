import React from "react"

import { Props } from "../component_interface"
import { emojize } from "../../util/emoji"

export function TText({ node }: Props) {
  return (
    <div id={node.props.id || undefined}>{emojize(node.props.text)}</div>
  )
}
