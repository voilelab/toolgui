import React, { useState } from "react"
import { NumberInput } from "@mantine/core"

import { stateValues } from "../state"
import { Props } from "../component_interface"
import { inputBorderStyles } from "../../util/color"

export function TNumber({ node, update }: Props) {
  // `??` rather than `||`: a stored 0 is a value the app user entered, and the
  // default only stands in until they have. Undefined when there is neither,
  // which is the empty box Mantine wants — not '', which would read as 0 and
  // trip the range check below.
  const [value, setValue] = useState<number>(
    stateValues[node.props.id] ?? node.props.default)

  const outOfRange =
    node.props.min !== undefined && value < node.props.min ||
    node.props.max !== undefined && value > node.props.max

  // What the app user left in the box goes back as it is, out of range or
  // not. Holding it back would leave the server on the last value that was in
  // range, and the page would read that as what is on screen now; sent, it is
  // the server that pulls it into the range.
  const send = () => update({
    type: "input",
    id: node.props.id,
    value: stateValues[node.props.id],
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
      value={value}
      onChange={(next) => {
        const val = Number(next)
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
