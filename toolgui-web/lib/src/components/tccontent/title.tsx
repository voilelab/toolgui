import React from "react"
import { Title } from "@mantine/core"

import { Props } from '../component_interface'
import { emojize } from '../../util/emoji'

export function TTitle({ node }: Props) {
  return (
    <Title id={node.props.id || undefined} order={1} mb="md">
      {emojize(node.props.text)}
    </Title>
  )
}
