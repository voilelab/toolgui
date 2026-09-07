import React from "react"
import { Button } from "@mantine/core"

import { Props } from "../component_interface"
import { mantineColor } from "../../util/color"

export function TButton({ node, update }: Props) {
  const color = mantineColor(node.props.color)

  return (
    <Button
      id={node.props.id}
      color={color}
      // A colourless button is neutral, the way Bulma's plain button is;
      // Mantine would otherwise fill it with the primary colour.
      variant={color ? 'filled' : 'default'}
      disabled={node.props.disabled}
      // The id comes from the node, not the event: Mantine wraps the label in
      // spans, so the click target is not always the <button>.
      onClick={() => {
        update({
          type: "click",
          id: node.props.id,
        })
      }}>
      {node.props.label}
    </Button>
  )
}
