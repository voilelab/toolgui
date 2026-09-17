import React, { useState } from "react"
import { NumberInput } from "@mantine/core"

import { stateValues } from "../state"
import { Props } from "../component_interface"
import { inputBorderStyles } from "../../util/color"

// What the box holds, with null for nothing at all.
type Value = number | null

// Mantine reports an empty box as '' and a half-typed number as a string, so
// neither is passed through Number() alone: '' would come back as 0 and '-'
// as NaN, and both would be sent as if the user had typed them.
function toValue(next: string | number): Value {
  if (next === '') {
    return null
  }

  const val = Number(next)
  return Number.isNaN(val) ? null : val
}

export function TNumber({ node, update }: Props) {
  // ?? rather than ||: a stored 0 is a value, not an absence.
  const [value, setValue] = useState<Value>(
    stateValues[node.props.id] ?? node.props.default ?? null)

  const outOfRange = value !== null && (
    node.props.min !== undefined && value < node.props.min ||
    node.props.max !== undefined && value > node.props.max)

  // What the user left in the box goes back as it is, out of range or empty.
  // Holding it back would leave the server on the last legal value, which the
  // page then reads as the current one.
  const send = () => update({
    type: "input",
    id: node.props.id,
    value: stateValues[node.props.id] ?? null,
  })

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
      value={value ?? ''}
      onChange={(next) => {
        const val = toValue(next)
        stateValues[node.props.id] = val
        setValue(val)
      }}
      onBlur={send}
      onKeyDown={(event) => {
        if (event.key === 'Enter') {
          send()
        }
      }}
    />
  )
}
