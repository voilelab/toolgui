import React from "react"
import { Alert } from "@mantine/core"

import { Props } from "../component_interface";
import { mantineColor } from "../../util/color";

export function TMessage({ node }: Props) {
  return (
    <Alert id={node.props.id || undefined}
      // A colourless message stays neutral, the way it always was; Mantine
      // would otherwise paint it in the primary colour.
      color={mantineColor(node.props.color) || 'gray'}
      title={node.props.title || undefined}
      mb="md">
      {node.props.body}
    </Alert>
  )
}
