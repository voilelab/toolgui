import React from "react"

import { Props } from "../component_interface"
import { TComponent } from "../factory"

const eqClass = [
  '', // 0
  'is-full',
  'is-half',
  'is-one-third',
  'is-one-quarter',
  'is-one-fifth',
]

export function TColumn({ node, update, upload, theme }: Props) {
  var columnClassname = 'column'
  if (node.props.equal) {
    columnClassname += ' ' + eqClass[node.children.length]
  }
  return (
    <div id={node.props.id} className="columns">
      {
        node.children.map(child =>
          <div key={child.props.id} className={columnClassname}>
            <TComponent node={child}
              update={update}
              upload={upload}
              theme={theme} />
          </div>
        )
      }
    </div>
  )
}
