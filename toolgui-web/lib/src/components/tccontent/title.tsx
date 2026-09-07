import React from "react"

import { Props } from '../component_interface'
import { emojize } from '../../util/emoji'

export function TTitle({ node }: Props) {
  return (
    <h1 id={node.props.id || undefined} className="title">
      {emojize(node.props.text)}
    </h1>
  )
}
