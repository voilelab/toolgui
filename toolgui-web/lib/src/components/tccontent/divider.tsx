import React from 'react'
import { Divider } from "@mantine/core"

import { Props } from "../component_interface"

export function TDivider({ node }: Props) {
  return (
    <Divider id={node.props.id || undefined} my="md" />
  )
}
