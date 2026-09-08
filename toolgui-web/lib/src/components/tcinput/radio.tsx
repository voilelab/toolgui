import React from "react";
import { Radio, Stack } from "@mantine/core"

import { stateValues } from "../state"
import { Props } from "../component_interface"

export function TRadio({ node, update }: Props) {
  const items: string[] = node.props.items

  // Unlike the select, 0 is a real item here, so the default is null rather
  // than 0 when there is none. It only stands in until the group is touched.
  const selected = stateValues[node.props.id] ?? node.props.default

  return (
    <Radio.Group
      id={node.props.id}
      value={selected == null ? null : String(selected)}
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
          <Radio key={idx} value={String(idx)} label={x}
            disabled={node.props.disabled} />
        )}
      </Stack>
    </Radio.Group>
  )
}
