import React from "react"
import { Paper } from "@mantine/core"

import { Props } from "../component_interface"
import { TComponent } from "../factory"

export function TBox({ node, update, upload, theme }: Props) {
  return (
    <Paper id={node.props.id || undefined}
      className="toolgui-box"
      shadow="xs" radius="md" p="md" mb="md">
      {
        node.children.map(child =>
          <TComponent key={child.reactKey} node={child}
            update={update}
            upload={upload}
            theme={theme} />
        )
      }
    </Paper>
  )
}
