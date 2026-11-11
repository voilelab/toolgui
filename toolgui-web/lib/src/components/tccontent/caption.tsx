import React from "react"
import { Text } from "@mantine/core"

import { Props } from "../component_interface"
import { emojize } from "../../util/emoji"

export function TCaption({ node }: Props) {
  return (
    <Text id={node.props.id || undefined} className="toolgui-caption"
      size="sm" c="dimmed">
      {emojize(node.props.text)}
    </Text>
  )
}
