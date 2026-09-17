import React from "react"

import { Props } from "../component_interface"
import { TComponent } from "../factory"

export function TContainer({ node, update, upload, download, theme }: Props) {
  return (
    <div id={node.props.id || undefined}>
      {
        node.children.map(child =>
          <TComponent key={child.reactKey} node={child}
            update={update}
            upload={upload}
            download={download}
            theme={theme} />
        )
      }
    </div>
  )
}