import React from 'react'
import { Title } from "@mantine/core"

import { Props } from '../component_interface'
import { emojize } from '../../util/emoji'
import '@toolgui-web/lib/src/assets/css/heading.css'

export function TSubtitle({ node }: Props) {
  return (
    <Title id={node.props.id || undefined} order={2} mb="sm"
      className="toolgui-subtitle">
      {emojize(node.props.text)}
    </Title>
  )
}
