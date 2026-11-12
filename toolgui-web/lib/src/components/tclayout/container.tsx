import React from "react"

import { Props } from "../component_interface"
import { TComponent } from "../factory"

import '@toolgui-web/lib/src/assets/css/stack.css'

export function TContainer({ node, update, upload, download, theme }: Props) {
  return (
    <div id={node.props.id || undefined} className="toolgui-stack">
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