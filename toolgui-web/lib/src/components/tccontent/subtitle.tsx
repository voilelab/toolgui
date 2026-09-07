import React from 'react'

import { Props } from '../component_interface'
import { emojize } from '../../util/emoji'

export function TSubtitle({ node }: Props) {
  return (
    <h2 id={node.props.id || undefined} className="subtitle">
      {emojize(node.props.text)}
    </h2>
  )
}
