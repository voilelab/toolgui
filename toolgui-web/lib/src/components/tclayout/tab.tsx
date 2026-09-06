import React, { useState } from 'react'
import { Props } from '../component_interface'
import { TComponent } from '../factory'

export function TTab({ node, update, upload, theme }: Props) {
  const [activeTab, setActiveTab] = useState(node.props.tabs[0])

  const activeIndex = node.props.tabs.indexOf(activeTab)

  if (node.props.tabs.length !== node.children.length) {
    // waiting for the children to be added by server
    return <></>
  }

  return (
    <>
      <div className="tabs">
        <ul>
          {
            node.props.tabs.map((tab: string) => (
              <li key={tab} className={`${activeTab === tab ? 'is-active' : ''} is-boxed`}>
                <a onClick={() => setActiveTab(tab)}>{tab}</a>
              </li>
            ))
          }
        </ul>
      </div>
      <div>
        <TComponent key={node.children[activeIndex].reactKey}
          node={node.children[activeIndex]}
          update={update}
          upload={upload}
          theme={theme} />
      </div>
    </>
  )
}
