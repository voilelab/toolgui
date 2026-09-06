import React from 'react'

import { Props } from "../component_interface"

export function TDivider({ node }: Props) {
  return (
    <hr id={node.props.id || undefined} />
  )
}
