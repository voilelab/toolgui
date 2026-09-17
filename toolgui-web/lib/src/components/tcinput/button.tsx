import React, { useContext } from "react"
import { Button } from "@mantine/core"

import { Props } from "../component_interface"
import { mantineColor } from "../../util/color"
import { FormSubmitContext } from "./form_context"

export function TButton({ node, update }: Props) {
  const color = mantineColor(node.props.color)

  // Null outside a form. Inside one, a button is what sends it.
  const submitForm = useContext(FormSubmitContext)

  return (
    <Button
      id={node.props.id}
      color={color}
      // A colourless button stays neutral; Mantine would otherwise fill it
      // with the primary colour.
      variant={color ? 'filled' : 'default'}
      disabled={node.props.disabled}
      // The id comes from the node, not the event: Mantine wraps the label in
      // spans, so the click target is not always the <button>.
      onClick={() => {
        update({
          type: "click",
          id: node.props.id,
        })

        // Inside a form the update above only queued the click. This sends it,
        // with the inputs queued before it, so the run that reports the click
        // is the one that reads the new values.
        submitForm?.()
      }}>
      {node.props.label}
    </Button>
  )
}
