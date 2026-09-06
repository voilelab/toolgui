import React, { useState } from "react"
import { stateValues } from "../state"
import { Props } from "../component_interface"

export function TTextarea({ node, update }: Props) {
  const [value, setValue] = useState<string>(stateValues[node.props.id] || node.props.default)

  var inputClassNames = 'textarea'
  if (node.props.color !== '') {
    inputClassNames += ' is-' + node.props.color
  }

  return (
    <div className="field">
      <label className="label">{node.props.label}</label>
      <div className="control">
        <textarea
          className={inputClassNames}
          id={node.props.id}
          value={value}
          rows={node.props.height}
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
          }}>
        </textarea>
      </div>
    </div>
  )
}
