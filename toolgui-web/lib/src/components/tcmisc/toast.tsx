import React, { useEffect, useRef } from "react"
import { notifications } from "@mantine/notifications"

import { Props } from "../component_interface"
import { emojize } from "../../util/emoji"

// TToast renders nothing. The toast it stands for goes to the <Notifications />
// container the app mounts and lives out its life there, so the node the page
// carries takes up no room and does not move what is around it.
//
// A toast is something that happened, and the page is a tree of nodes. The
// same Toast call on the next run lands on the same node, so a mount is not
// what says "fire": the run serial the server puts in the props is. It is the
// one prop that differs between two runs that wrote the same toast.
export function TToast({ node }: Props) {
  const seq: number = node.props.seq

  // The run this toast was last fired for. Held in a ref rather than left to
  // the effect's dependencies alone, so that React re-running the effect on
  // the same props -- StrictMode does exactly this in development -- does not
  // show the toast twice.
  const firedSeq = useRef<number | null>(null)

  useEffect(() => {
    if (firedSeq.current === seq) {
      return
    }
    firedSeq.current = seq

    const icon: string = node.props.icon
    const duration: number = node.props.duration_ms

    notifications.show({
      message: emojize(node.props.text),
      // A duration of zero is "however long the container says".
      autoClose: duration > 0 ? duration : undefined,
      icon: icon ? <span>{emojize(icon)}</span> : undefined,
      // The component's id names the call site, and so is the same on every
      // run. It is put on the element rather than handed to Mantine as the
      // notification id: showing an id that is already on screen is a no-op
      // there, and the second toast of a repeated run would be swallowed.
      // Mantine gives each notification an id of its own instead.
      "data-toolgui-id": node.props.id || undefined,
    })
    // The rest of the props belong to the run that sent this serial, so the
    // serial alone decides when this runs again.
  }, [seq])

  // Nothing is drawn where the page wrote the toast; the notification itself
  // is rendered by the <Notifications /> container, over the page.
  return <></>
}
