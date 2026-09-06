import React from "react"

import { Props } from "../component_interface"
import { TComponent } from "../factory"

import { UpdateEvent } from "../../app/UpdateEvent"

export function TForm({ node, update, upload, theme }: Props) {
  const collectEvent: UpdateEvent[] = []

  const handleUpdate = (event: UpdateEvent) => {
    collectEvent.push(event)
  }

  return (
    <div id={node.props.id || undefined}>
      <div className="field">
        {
          node.children.map(child =>
            <TComponent key={child.reactKey} node={child}
              update={handleUpdate}
              upload={upload}
              theme={theme} />
          )
        }
      </div>

      <div className="field">
        <p className="control">
          <button onClick={() => {
            update({
              type: "form",
              events: collectEvent,
            })

            collectEvent.splice(0, collectEvent.length)
          }} className="button">Submit</button>
        </p>
      </div>
    </div>
  )
}
