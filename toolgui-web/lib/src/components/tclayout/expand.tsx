import React, { useState } from "react"

import { Props } from "../component_interface"
import { TComponent } from "../factory"

export function TExpand({ node, update, upload, theme }: Props) {
  const [expanded, setExpanded] = useState(node.props.expanded)
  return (
    <div className="card">
      <header className="card-header" onClick={() => setExpanded(!expanded)}>
        <div className="card-header-icon">
          <span className="icon">
            <i className={`fas ${expanded ? 'fa-minus' : 'fa-plus'}`}></i>
          </span>
        </div>
        <p className="card-header-title">{node.props.title}</p>
      </header>
      {expanded &&
        <div className="card-content">
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