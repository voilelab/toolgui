import React, { useState } from "react"
import { MultiSelect } from "@mantine/core"

import { stateValues } from "../state"
import { Props } from "../component_interface"

export function TMultiselect({ node, update }: Props) {
  const items: string[] = node.props.items
  const max: number = node.props.max_selections || 0

  // Nothing is selected until the component is first touched, which is when
  // the default stands in — the same rule Go applies to the state. Held in
  // React too, so the pills follow the click inside a form, where no answer
  // comes back from the server to re-render on.
  const [selected, setSelected] = useState<number[]>(
    stateValues[node.props.id] || node.props.default || [])

  const atMax = max > 0 && selected.length >= max

  // The options are numbered rather than labelled, so duplicate labels stay
  // apart and what the server reads is already an index.
  const data = items.map((item: string, index: number) => ({
    value: String(index),
    label: item,
    // At the cap the rest go disabled, so the limit shows in the dropdown
    // rather than arriving as a refusal after the click.
    disabled: atMax && !selected.includes(index),
  }))

  return (
    <MultiSelect
      id={node.props.id}
      label={node.props.label}
      placeholder={node.props.placeholder}
      disabled={node.props.disabled}
      data={data}
      mb="md"
      value={selected.map(String)}
      onChange={(values) => {
        // Ordered by items rather than by the order they were picked in,
        // which is what the Go side hands back.
        const indices = values.map(Number).sort((a, b) => a - b)
        stateValues[node.props.id] = indices
        setSelected(indices)
        update({
          type: "select",
          id: node.props.id,
          values: indices,
        })
      }} />
  )
}
