import React, { useRef } from "react"
import { Button } from "@mantine/core"

import { Props } from "../component_interface"
import { TComponent } from "../factory"
import { stateGeneration } from "../state"
import { FormSubmitContext } from "./form_context"

import { UpdateEvent } from "../../app/UpdateEvent"

export function TForm({ node, update, upload, download, theme }: Props) {
  // A ref, not a plain const: the queue has to survive a re-render. Nothing
  // inside a form reruns the page, so a queue rebuilt by the next render of
  // whatever is around the form would take the held values with it.
  const collectEvent = useRef<UpdateEvent[]>([])

  // Outliving a render also means outliving a server session reset, which
  // clears everything else the client holds. Those values belong to a session
  // that is gone, so the queue is dropped rather than replayed into the new
  // one.
  const generation = useRef(stateGeneration)
  const queue = () => {
    if (generation.current !== stateGeneration) {
      generation.current = stateGeneration
      collectEvent.current = []
    }

    return collectEvent.current
  }

  const submit = () => {
    update({
      type: "form",
      events: queue(),
    })

    collectEvent.current = []
  }

  const handleUpdate = (event: UpdateEvent) => {
    queue().push(event)
  }

  return (
    <div id={node.props.id || undefined}>
      <FormSubmitContext.Provider value={submit}>
        {
          node.children.map(child =>
            <TComponent key={child.reactKey} node={child}
              update={handleUpdate}
              upload={upload}
              download={download}
              theme={theme} />
          )
        }
      </FormSubmitContext.Provider>

      {!node.props.hide_submit &&
        <Button variant="default" onClick={submit}>
          {node.props.submit_label || "Submit"}
        </Button>}
    </div>
  )
}
