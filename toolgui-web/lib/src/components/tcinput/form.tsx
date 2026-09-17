import React, { useRef } from "react"
import { Button } from "@mantine/core"

import { Props } from "../component_interface"
import { TComponent } from "../factory"

import { UpdateEvent } from "../../app/UpdateEvent"

export function TForm({ node, update, upload, theme }: Props) {
  // A ref, not a plain const: the queue has to survive a re-render. Nothing
  // inside a form reruns the page, so a queue rebuilt by the next render of
  // whatever is around the form would take the held values with it.
  const collectEvent = useRef<UpdateEvent[]>([])

  const submit = () => {
    update({
      type: "form",
      events: collectEvent.current,
    })

    collectEvent.current = []
  }

  const handleUpdate = (event: UpdateEvent) => {
    collectEvent.current.push(event)

    // A click inside a form submits it, so a form can carry its own button
    // instead of a second, hardwired one. The click goes last, after the
    // inputs it was made with, which is the order the server replays them in.
    if (event.type === "click") {
      submit()
    }
  }

  return (
    <div id={node.props.id || undefined}>
      {
        node.children.map(child =>
          <TComponent key={child.reactKey} node={child}
            update={handleUpdate}
            upload={upload}
            theme={theme} />
        )
      }

      {!node.props.hide_submit &&
        <Button variant="default" onClick={submit}>
          {node.props.submit_label || "Submit"}
        </Button>}
    </div>
  )
}
