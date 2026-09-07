import React from "react"
import { Button } from "@mantine/core"

import { Props } from '../component_interface'
import { mantineColor } from "../../util/color"

export function TDownloadButton({ node, update }: Props) {
  const color = mantineColor(node.props.color)

  return (
    <Button id={node.props.id}
      color={color}
      variant={color ? 'filled' : 'default'}
      disabled={node.props.disabled}
      onClick={() => {
        var link = document.createElement('a')
        link.setAttribute('download', node.props.filename)
        link.href = node.props.uri
        document.body.appendChild(link)
        link.click()
        document.body.removeChild(link)
        URL.revokeObjectURL(node.props.uri)
        update({
          type: "click",
          id: node.props.id,
        })
      }}>
      {node.props.text}
    </Button>
  )
}
