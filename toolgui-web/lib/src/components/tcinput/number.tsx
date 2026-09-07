import React, { useState } from "react"
import { NumberInput } from "@mantine/core"

import { stateValues } from "../state"
import { Props } from "../component_interface"
import { inputBorderStyles } from "../../util/color"

export function TNumber({ node, update }: Props) {
  const [value, setValue] = useState<number>(stateValues[node.props.id] || node.props.default)

  const outOfRange =
    node.props.min !== undefined && value < node.props.min ||
    node.props.max !== undefined && value > node.props.max

  return (
    <NumberInput
      id={node.props.id}
      label={node.props.label}
      placeholder={node.props.placeholder}
      disabled={node.props.disabled}
      min={node.props.min}
      max={node.props.max}
      step={node.props.step}
      // The range is reported, not enforced: Mantine would otherwise pull the
      // value back into it and the message would never be seen.
      clampBehavior="none"
      styles={inputBorderStyles(node.props.color)}
      error={outOfRange ? 'Value out of range' : undefined}
      mb="md"
      value={value}
      onChange={(next) => {
        const val = Number(next)
        stateValues[node.props.id] = val
        setValue(val)
      }}
      onBlur={() => {
        const val = stateValues[node.props.id]

        if (node.props.min !== undefined && val < node.props.min) {
          return
        }

        if (node.props.max !== undefined && val > node.props.max) {
          return
        }

        update({
          type: "input",
          id: node.props.id,
          value: val,
        })
      }}
    />
  )
}
