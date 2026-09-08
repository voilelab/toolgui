import React from "react"
import { Select } from "@mantine/core"

import { stateValues } from "../state"
import { Props } from "../component_interface"

export function TSelect({ node, update }: Props) {
  // Go numbers the items from 1 and keeps 0 for "nothing selected", so the
  // option values are those indices rather than the labels — which also keeps
  // duplicate labels apart.
  const data = node.props.items.map((item: string, index: number) => ({
    value: String(index + 1),
    label: item,
  }))

  // The default is 1-based too, so it drops straight into the same slot, and
  // only stands in until the component is first touched.
  const selected = stateValues[node.props.id] ?? node.props.default ?? 0

  return (
    <Select
      id={node.props.id}
      label={node.props.label}
      placeholder="Please select an option"
      disabled={node.props.disabled}
      data={data}
      value={selected > 0 ? String(selected) : null}
      onChange={(value) => {
        const index = value === null ? 0 : Number(value)
        stateValues[node.props.id] = index
        update({
          type: "select",
          id: node.props.id,
          value: index,
        })
      }} />
  )
}
