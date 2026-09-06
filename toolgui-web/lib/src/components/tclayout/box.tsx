import React from "react"

import { Props } from "../component_interface"
import { TComponent } from "../factory"

export function TBox({ node, update, upload, theme }: Props) {
  return (
    <div id={node.props.id} className="box">
      {
        node.children.map(child =>
          <TComponent key={child.props.id} node={child}
            update={update}
            upload={upload}
            theme={theme} />
        )
      }
    </div>
  )
}
