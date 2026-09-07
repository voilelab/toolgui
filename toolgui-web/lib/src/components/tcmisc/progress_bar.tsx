import React from "react"
import { Progress, Text } from "@mantine/core"

import { Props } from "../component_interface";

export function TProgressar({ node }: Props) {
  return (
    <div>
      <Text>{node.props.label}</Text>
      <Progress id={node.props.id || undefined}
        value={node.props.value} mb="md" />
    </div>
  )
}
