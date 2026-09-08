import React, { useState } from "react"
import { Input, Slider } from "@mantine/core"

import { stateValues } from "../state"
import { Props } from "../component_interface"

export function TSlider({ node, update }: Props) {
  const id = node.props.id
  const [value, setValue] = useState<number>(
    stateValues[id] ?? node.props.default)

  return (
    // Mantine's Slider takes no label of its own — its `label` is the value
    // bubble — so the wrapper supplies one. It labels a div rather than a
    // form control, hence labelElement.
    <Input.Wrapper label={node.props.label} labelElement="div" mb="md">
      <Slider
        id={id}
        min={node.props.min}
        max={node.props.max}
        step={node.props.step}
        disabled={node.props.disabled}
        value={value}
        // A drag crosses every step on the way and each one would be a
        // server-side rerun, so the handle moves locally and the value is
        // only reported once it is let go.
        onChange={setValue}
        onChangeEnd={(next) => {
          stateValues[id] = next
          update({
            type: "input",
            id: id,
            value: next,
          })
        }} />
    </Input.Wrapper>
  )
}
