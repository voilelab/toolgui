import React, { useState } from "react"
import { Checkbox } from "@mantine/core"

import { stateValues } from "../state"
import { Props } from "../component_interface"

export function TCheckbox({ node, update }: Props) {
  const id = node.props.id

  // Controlled input: keep the value in React state, or the tick reverts
  // until the rerun lands (and never moves inside a form). Same as TToggle.
  const [checked, setChecked] = useState<boolean>(
    stateValues[id] ?? node.props.default)

  return (
    <Checkbox
      id={id}
      label={node.props.label}
      checked={checked}
      disabled={node.props.disabled}
      mb="md"
      onChange={(event) => {
        const next = event.currentTarget.checked
        stateValues[id] = next
        setChecked(next)
        update({
          type: "input",
          id: id,
          value: next,
        })
      }} />
  )
}
