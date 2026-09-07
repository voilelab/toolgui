import React from "react"
import { Anchor } from "@mantine/core"

import { Props } from '../component_interface'
import { emojize } from '../../util/emoji'

export function TLink({ node }: Props) {
  return (
    <Anchor id={node.props.id || undefined} href={node.props.url}>
      {emojize(node.props.text)}
    </Anchor>
  )
}
