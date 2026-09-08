import React from "react"
import { DateInput, DateTimePicker, TimeInput } from "@mantine/dates"
import dayjs from "dayjs"

import { stateValues } from "../state"
import { Props } from "../component_interface"

// What Go's time.Parse expects back, per picker type. Mantine's own value
// format is a display concern and is set separately.
const DATETIME_WIRE_FORMAT = 'YYYY-MM-DDTHH:mm'

// Mantine reports a datetime as 'YYYY-MM-DD HH:mm:ss'.
const DATETIME_MANTINE_FORMAT = 'YYYY-MM-DD HH:mm:ss'

export function TDatepicker({ node, update }: Props) {
  const id = node.props.id
  // The default only stands in until the picker is first touched, and an
  // empty string is a value the app user cleared rather than one to fill in.
  const stored = stateValues[id] ?? node.props.default ?? ''

  const send = () => {
    update({
      type: "input",
      id: id,
      value: stateValues[id],
    })
  }

  if (node.props.type === 'time') {
    // Still a native <input type="time">, Mantine only styles it, so the
    // browser hands back the HH:mm Go wants.
    return (
      <TimeInput
        id={id}
        label={node.props.label}
        disabled={node.props.disabled}
        defaultValue={stored}
        onChange={(event) => { stateValues[id] = event.currentTarget.value }}
        onBlur={send} />
    )
  }

  if (node.props.type === 'datetime-local') {
    // The value only moves on an explicit pick, so there is no half-typed
    // state to wait out — send it as it changes.
    return (
      <DateTimePicker
        id={id}
        label={node.props.label}
        disabled={node.props.disabled}
        valueFormat="YYYY-MM-DD HH:mm"
        defaultValue={stored ?
          dayjs(stored).format(DATETIME_MANTINE_FORMAT) : null}
        onChange={(value) => {
          stateValues[id] = value ?
            dayjs(value).format(DATETIME_WIRE_FORMAT) : ''
          send()
        }} />
    )
  }

  // Typing reports every keystroke that parses as a date, so hold the value
  // until blur — the same point the other text inputs report at.
  return (
    <DateInput
      id={id}
      label={node.props.label}
      disabled={node.props.disabled}
      valueFormat="YYYY-MM-DD"
      defaultValue={stored || null}
      onChange={(value) => { stateValues[id] = value || '' }}
      onBlur={send} />
  )
}
