import React from "react"

import { Props } from '../component_interface'

export function TDownloadButton({ node, update }: Props) {
  var className = 'button'

  if (node.props.color) {
    className += ' is-' + node.props.color
  }

  return (
    <button id={node.props.id}
      className={className}
      onClick={
        (event: React.MouseEvent<HTMLButtonElement>) => {
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
        }
      }
      disabled={node.props.disabled}
    >
      {node.props.text}
    </button>
  )
}
