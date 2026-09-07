import React from "react"
import { Grid, GridColProps } from "@mantine/core"

import { Props } from "../component_interface"
import { TComponent } from "../factory"

export function TColumn({ node, update, upload, theme }: Props) {
  const count = node.children.length

  // equal splits the row into one share per child; otherwise a column takes
  // the space its content needs and shares out the rest. Below the tablet
  // breakpoint a column is full width, the way it always was.
  const columns = node.props.equal ? Math.max(count, 1) : undefined
  const span: GridColProps['span'] = node.props.equal ?
    { base: Math.max(count, 1), sm: 1 } :
    { base: 12, sm: 'auto' }

  return (
    <Grid id={node.props.id || undefined} columns={columns}>
      {
        node.children.map(child =>
          // min-width:0 keeps wide content (code blocks, tables) scrolling
          // inside the column instead of stretching it past the row.
          <Grid.Col key={child.reactKey} span={span} style={{ minWidth: 0 }}>
            <TComponent node={child}
              update={update}
              upload={upload}
              theme={theme} />
          </Grid.Col>
        )
      }
    </Grid>
  )
}
