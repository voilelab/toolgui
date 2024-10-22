import React from "react"
import Button from '@mui/material/Button'

import { Props } from "../component_interface"

export function TButton({ node, update }: Props) {
  var className = 'button'

  if (node.props.color) {
    className += ' is-' + node.props.color
  }

  return (
    <Button
      variant="contained"
      color={node.props.color}
      id={node.props.id}
      disabled={node.props.disabled}
      onClick={
        (event: React.MouseEvent<HTMLButtonElement>) => {
          const target = event.target as HTMLButtonElement
          update({
            type: "click",
            id: target.id,
          })
        }
      }>
      {node.props.label}
    </Button>
  )
}
