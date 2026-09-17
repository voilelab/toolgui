import React from "react"

import { Props } from "../component_interface"
import { TComponent } from "../factory"

// TEmpty is the place an Empty slot writes into. It draws nothing of its own:
// the slot is whatever the page last wrote there, and nothing when the page
// cleared it.
export function TEmpty({ node, update, upload, download, theme }: Props) {
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
