import React from "react"
import { Alert } from "@mantine/core"

import { Props } from "../component_interface";

// TError is the placeholder a component leaves behind when the run's data
// stopped it from being drawn. The run carries on, so this marks the one spot
// that failed while the rest of the page renders around it.
export function TError({ node }: Props) {
  return (
    <Alert id={node.props.id || undefined}
      color="red"
      title="Component error"
      mb="md">
      {node.props.message}
    </Alert>
  )
}
