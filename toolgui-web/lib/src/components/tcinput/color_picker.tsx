import React, { useState } from "react"
import { ColorInput } from "@mantine/core"

import { stateValues } from "../state"
import { Props } from "../component_interface"

export function TColorPicker({ node, update }: Props) {
  const id = node.props.id
  const [color, setColor] = useState<string>(
    stateValues[id] ?? node.props.default)

  return (
    <ColorInput
      id={id}
      label={node.props.label}
      // Go is typed against "#rrggbb", so the picker is held to that format.
      format="hex"
      disabled={node.props.disabled}
      mb="md"
      value={color}
      // Dragging over the saturation area reports continuously; the value
      // follows locally and is only reported once the pointer is let go.
      onChange={setColor}
      onChangeEnd={(next) => {
        const value = next.toLowerCase()
        stateValues[id] = value
        update({
          type: "input",
          id: id,
          value: value,
        })
      }} />
  )
}
