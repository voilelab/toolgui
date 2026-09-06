import React, { useState } from "react"

import { Props } from "../component_interface"
import { TComponent } from "../factory"

export function TExpand({ node, update, upload, theme }: Props) {
  const [expanded, setExpanded] = useState(node.props.expanded)
  // Contents are built on the first open and then kept mounted, so an
  // expander that was never opened costs nothing and collapsing one that
  // was opened does not throw away what it holds.
  const [everExpanded, setEverExpanded] = useState(node.props.expanded)

  const toggle = () => {
    setExpanded(!expanded)
    setEverExpanded(true)
  }

  return (
    <div className="card">
      <header className="card-header" onClick={toggle}>
        <div className="card-header-icon">
          <span className="icon">
            <i className={`fas ${expanded ? 'fa-minus' : 'fa-plus'}`}></i>
          </span>
        </div>
        <p className="card-header-title">{node.props.title}</p>
      </header>
      {everExpanded &&
        <div className="card-content" hidden={!expanded}>
          <div className="content">
            {
              node.children.map(child =>
                <TComponent key={child.reactKey} node={child}
                  update={update}
                  upload={upload}
                  theme={theme} />
              )
            }
          </div>
        </div>
      }
    </div>
  )
}
