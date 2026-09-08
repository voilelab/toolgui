import React, { useState } from "react"
import { Switch } from "@mantine/core"

import { stateValues } from "../state"
import { Props } from "../component_interface"

export function TToggle({ node, update }: Props) {
  const id = node.props.id

  // A Switch is a controlled input, so the flip has to move a piece of React
  // state or the next render pulls it straight back. Outside a form the rerun
  // hides that -- the tree comes back with the value applied -- but inside
  // one nothing reruns until Submit, and the switch would sit there refusing
  // to move.
  const [checked, setChecked] = useState<boolean>(
    stateValues[id] ?? node.props.default)

  return (
    <Switch
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
