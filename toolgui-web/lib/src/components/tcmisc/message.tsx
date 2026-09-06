import React from "react"

import { Props } from "../component_interface";

export function TMessage({ node }: Props) {
  var className = 'message'

  if (node.props.color) {
    className += ' is-' + node.props.color
  }

  return (
    <article id={node.props.id || undefined} className={className}>
      {node.props.title ?
        <div className="message-header">
          <p>{node.props.title}</p>
        </div> : ''}
      <div className="message-body">
        {node.props.body}
      </div>
    </article>
  )
}

