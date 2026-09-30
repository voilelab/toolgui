import React, { useState } from "react"
import { Textarea } from "@mantine/core"

import { stateValues, syncResetKey } from "../state"
import { Props } from "../component_interface"
import { inputBorderStyles } from "../../util/color"

// A new reset_key remounts the input, which drops the typed value.
export function TTextarea(props: Props) {
  return <TextareaInner key={props.node.props.reset_key ?? ''} {...props} />
}

function TextareaInner({ node, update }: Props) {
  const [value, setValue] = useState<string>(() => {
    syncResetKey(node.props.id, node.props.reset_key)
    return stateValues[node.props.id] || node.props.default
  })

  return (
    <Textarea
      id={node.props.id}
      label={node.props.label}
      rows={node.props.height}
      styles={inputBorderStyles(node.props.color)}
      mb="md"
      value={value}
      onChange={(event) => {
        stateValues[event.target.id] = event.target.value
        setValue(event.target.value)
      }}
      onBlur={(event) => {
        update({
          type: "input",
          id: event.target.id,
          value: stateValues[event.target.id],
        })
      }}
    />
  )
}
