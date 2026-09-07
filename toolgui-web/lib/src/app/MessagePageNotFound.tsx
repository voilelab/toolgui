import React from "react"
import { Alert } from "@mantine/core"

export function MessagePageNotFound() {
  return (
    <Alert color="yellow" title="Oops! Page not found."
      maw="75%" mx="auto">
      We're sorry, the page you requested was not found.
      Try using the navigation menu to find what you're looking for.
    </Alert>
  )
}
