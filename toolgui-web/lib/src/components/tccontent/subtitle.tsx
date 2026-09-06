import React from 'react'

import { Props } from '../component_interface'

export function TSubtitle({ node }: Props) {
  return (
    <h2 id={node.props.id || undefined} className="subtitle">
      {node.props.text}
    </h2>
  )
}
