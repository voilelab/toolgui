import React, { useState } from "react"
import { Input, Slider } from "@mantine/core"

import { stateValues } from "../state"
import { Props } from "../component_interface"

export function TSelectSlider({ node, update }: Props) {
  const id = node.props.id
  const items: string[] = node.props.items
  const [index, setIndex] = useState<number>(
    stateValues[id] ?? node.props.default)

  // The slider runs over the item indices, so every position is an item and
  // the labels come from the items rather than from the numbers.
  const marks = items.map((item, idx) => ({ value: idx, label: item }))

  return (
    // The marks are drawn under the track, so the bottom margin is the wider
    // one — see slider.tsx for the wrapper itself.
    <Input.Wrapper label={node.props.label} labelElement="div" mb="xl">
      <Slider
        id={id}
        min={0}
        max={items.length - 1}
        step={1}
        marks={marks}
        label={(value) => items[value]}
        disabled={node.props.disabled}
        value={index}
        // Reported on release, for the reason given in slider.tsx.
        onChange={setIndex}
        onChangeEnd={(next) => {
          stateValues[id] = next
          update({
            type: "select",
            id: id,
            value: next,
          })
        }} />
    </Input.Wrapper>
  )
}
