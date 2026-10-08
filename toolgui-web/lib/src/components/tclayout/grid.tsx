import React from "react"

import { Props } from "../component_interface"
import { TComponent } from "../factory"

import '@toolgui-web/lib/src/assets/css/grid.css'

export function TGrid({ node, update, upload, download, theme }: Props) {
  const minColWidth: number = node.props.min_col_width || 200
  const gap: string = node.props.gap || "md"

  // auto-fill keeps empty tracks, so a short last row lines up with the rows
  // above; min() lets a single cell shrink below minColWidth on a narrow
  // screen instead of overflowing it.
  const style = {
    gridTemplateColumns:
      `repeat(auto-fill, minmax(min(${minColWidth}px, 100%), 1fr))`,
    gap: gap === "none" ? 0 : `var(--mantine-spacing-${gap})`,
  }

  return (
    <div id={node.props.id || undefined} className="toolgui-grid" style={style}>
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
