import React, { useState } from "react"
import { stateValues } from "../state"
import { Props } from "../component_interface"


export function TNumber({ node, update }: Props) {
  const [value, setValue] = useState<number>(stateValues[node.props.id] || node.props.default)

  var inputClassNames = 'input'
  if (node.props.color !== '') {
    inputClassNames += ' is-' + node.props.color
  }

  var errorDOM: React.ReactNode
  if (node.props.min !== undefined && value < node.props.min || node.props.max !== undefined && value > node.props.max) {
    errorDOM = <p className="help is-danger">Value out of range</p>
  }

  return (
    <div className="field">
      <label className="label">{node.props.label}</label>
      <div className="control">
        <input
          type="number"
          id={node.props.id}
          className={inputClassNames}
          placeholder={node.props.placeholder}
          disabled={node.props.disabled}
          min={node.props.min}
          max={node.props.max}
          step={node.props.step}
          value={value}
          onChange={(event) => {
            const val = Number(event.target.value)
            stateValues[event.target.id] = val
            setValue(val)
          }}
          onBlur={(event) => {
            const val = stateValues[event.target.id]

            if (node.props.min !== undefined && val < node.props.min) {
              return
            }

            if (node.props.max !== undefined && val > node.props.max) {
              return
            }

            update({
              type: "input",
              id: event.target.id,
              value: val,
            })
          }}
        />
      </div>
      {errorDOM}
    </div>
  )
}
