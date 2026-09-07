import React from "react";
import { Radio, Stack } from "@mantine/core"

import { stateValues } from "../state"
import { Props } from "../component_interface"

export function TRadio({ node, update }: Props) {
  const items: string[] = node.props.items
  const selected = stateValues[node.props.id]

  return (
    <Radio.Group
      id={node.props.id}
      value={selected === undefined ? null : String(selected)}
      mb="md"
      onChange={(value) => {
        stateValues[node.props.id] = Number(value)
        update({
          type: "select",
          id: node.props.id,
          value: Number(value),
        })
      }}>
      <Stack gap="xs">
        {items.map((x, idx) =>
          <Radio key={idx} value={String(idx)} label={x} />
        )}
      </Stack>
    </Radio.Group>
  )
}
