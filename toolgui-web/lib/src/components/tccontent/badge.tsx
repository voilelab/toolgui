import React from "react"
import { Badge } from "@mantine/core"

import { Props } from "../component_interface"
import { mantineColor } from "../../util/color"
import { emojize } from "../../util/emoji"

export function TBadge({ node }: Props) {
  const color = mantineColor(node.props.color)

  return (
    <Badge id={node.props.id || undefined}
      color={color}
      // A colourless badge stays neutral; Mantine would otherwise fill it
      // with the primary colour.
      variant={color ? 'filled' : 'default'}>
      {emojize(node.props.text)}
    </Badge>
  )
}
