import React, { useState } from "react"
import { Text, TextInput } from "@mantine/core"

import { stateValues, syncResetKey } from "../state"
import { Props } from "../component_interface"
import { inputBorderStyles } from "../../util/color"

// A new reset_key remounts the input, which drops the typed value.
export function TTextbox(props: Props) {
  return <TextboxInner key={props.node.props.reset_key ?? ''} {...props} />
}

function TextboxInner({ node, update }: Props) {
  const [value, setValue] = useState<string>(() => {
    syncResetKey(node.props.id, node.props.reset_key)
    return stateValues[node.props.id] || node.props.default
  })

  return (
    <TextInput
      type={node.props.password ? 'password' : 'text'}
      id={node.props.id}
      label={node.props.label}
      maxLength={node.props.max_length ? node.props.max_length : undefined}
      placeholder={node.props.placeholder}
      disabled={node.props.disabled}
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
      inputWrapperOrder={['label', 'input', 'description']}
      description={node.props.max_length > 0 ?
        <Text component="span" size="xs" ta="right" display="block">
          {value.length}/{node.props.max_length}
        </Text> : undefined}
    />
  )
}
