import React from "react"
import { Group, Loader, Text } from "@mantine/core"

import { Props } from "../component_interface"

export function TSpinner({ node }: Props) {
  return (
    <Group id={node.props.id || undefined} gap="xs" mb="md">
      <Loader size="sm" />
      <Text>{node.props.label}</Text>
    </Group>
  )
}
