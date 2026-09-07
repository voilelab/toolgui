import React from 'react'
import { Title } from "@mantine/core"

import { Props } from '../component_interface'
import { emojize } from '../../util/emoji'

export function TSubtitle({ node }: Props) {
  return (
    <Title id={node.props.id || undefined} order={2} mb="sm">
      {emojize(node.props.text)}
    </Title>
  )
}
