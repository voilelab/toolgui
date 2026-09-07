import React from "react"

import { Props } from '../component_interface'
import { emojize } from '../../util/emoji'

export function TLink({ node }: Props) {
  return (
    <a id={node.props.id || undefined} href={node.props.url}>
      {emojize(node.props.text)}
    </a>
  )
}
