import React from "react"
import { Group, GroupProps } from "@mantine/core"

import { Props } from "../component_interface"
import { TComponent } from "../factory"

import '@toolgui-web/lib/src/assets/css/toolbar.css'

// The flex alignment for each of the ways the server names.
const JUSTIFY: { [justify: string]: GroupProps['justify'] } = {
  start: "flex-start",
  end: "flex-end",
  between: "space-between",
}

export function TToolbar({ node, update, upload, download, theme }: Props) {
  const sticky = node.props.sticky

  return (
    <Group id={node.props.id || undefined}
      className={sticky ? "toolgui-toolbar is-sticky" : "toolgui-toolbar"}
      justify={JUSTIFY[node.props.justify] ?? JUSTIFY.start}
      // A row of controls, so they sit tight together and wrap rather than
      // push the page wider than the viewport.
      gap="xs" wrap="wrap">
      {
        node.children.map(child =>
          <TComponent key={child.reactKey} node={child}
            update={update}
            upload={upload}
            download={download}
            theme={theme} />
        )
      }
    </Group>
  )
}
