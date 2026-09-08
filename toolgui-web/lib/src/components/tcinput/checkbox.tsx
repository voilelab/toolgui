import React from "react"
import { Checkbox } from "@mantine/core"

import { stateValues } from "../state"
import { Props } from "../component_interface"

export function TCheckbox({ node, update }: Props) {
  return (
    <Checkbox
      id={node.props.id}
      label={node.props.label}
      checked={stateValues[node.props.id] ?? node.props.default}
      disabled={node.props.disabled}
      onChange={(event) => {
        const checked = event.currentTarget.checked
        stateValues[node.props.id] = checked
        update({
          type: "input",
          id: node.props.id,
          value: checked,
        })
      }} />
  )
}
