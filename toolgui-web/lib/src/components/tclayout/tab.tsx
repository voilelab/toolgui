import React, { useState } from 'react'
import { Tabs } from '@mantine/core'

import { Props } from '../component_interface'
import { TComponent } from '../factory'

export function TTab({ node, update, upload, theme }: Props) {
  const [activeTab, setActiveTab] = useState<string | null>(node.props.tabs[0])

  if (node.props.tabs.length !== node.children.length) {
    // waiting for the children to be added by server
    return <></>
  }

  return (
    // Every tab stays mounted and the inactive ones are hidden, so switching
    // away does not throw away what a tab holds.
    <Tabs id={node.props.id || undefined}
      value={activeTab} onChange={setActiveTab}
      keepMounted keepMountedMode="display-none">
      <Tabs.List>
        {
          node.props.tabs.map((tab: string) => (
            <Tabs.Tab key={tab} value={tab}>{tab}</Tabs.Tab>
          ))
        }
      </Tabs.List>

      {
        node.children.map((child, index) => (
          <Tabs.Panel key={child.reactKey} value={node.props.tabs[index]} pt="sm">
            <TComponent node={child}
              update={update}
              upload={upload}
              theme={theme} />
          </Tabs.Panel>
        ))
      }
    </Tabs>
  )
}
