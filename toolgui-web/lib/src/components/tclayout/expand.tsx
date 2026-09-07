import React, { useState } from "react"
import { Accordion } from "@mantine/core"

import { Props } from "../component_interface"
import { TComponent } from "../factory"

const ITEM = 'expand'

export function TExpand({ node, update, upload, theme }: Props) {
  const [expanded, setExpanded] = useState(node.props.expanded)
  // Contents are built on the first open and then kept mounted, so an
  // expander that was never opened costs nothing and collapsing one that
  // was opened does not throw away what it holds.
  const [everExpanded, setEverExpanded] = useState(node.props.expanded)

  return (
    <Accordion variant="contained" chevronPosition="left"
      value={expanded ? ITEM : null}
      onChange={(value) => {
        setExpanded(value === ITEM)
        if (value === ITEM) {
          setEverExpanded(true)
        }
      }}>
      <Accordion.Item value={ITEM}>
        <Accordion.Control>{node.props.title}</Accordion.Control>
        <Accordion.Panel keepMountedMode="display-none">
          {everExpanded &&
            node.children.map(child =>
              <TComponent key={child.reactKey} node={child}
                update={update}
                upload={upload}
                theme={theme} />
            )}
        </Accordion.Panel>
      </Accordion.Item>
    </Accordion>
  )
}
