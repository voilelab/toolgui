import React, { useEffect, useRef, useState } from "react"
import { Modal } from "@mantine/core"

import { Props } from "../component_interface"
import { TComponent } from "../factory"

// Mantine's size for each of the widths the server names.
const SIZES: { [width: string]: string } = {
  small: "sm",
  medium: "md",
  large: "lg",
}

// openDialogs is every dialog currently on screen, in the order they opened,
// which is also the order they are stacked in. Only the last one answers ESC:
// Mantine's own closeOnEscape is a window handler per Modal with no notion of
// a stack, so leaving it on closes every open dialog on one press.
const openDialogs: string[] = []

// Mantine components that handle ESC themselves, a Select with its dropdown
// open among them, mark the event this way. Same check Mantine's own modal
// makes, so a dropdown inside a dialog still closes on its own first.
function handledInside(target: EventTarget | null): boolean {
  return target instanceof Element &&
    target.getAttribute("data-mantine-stop-propagation") === "true"
}

export function TDialog({ node, update, upload, theme }: Props) {
  // Closing is shown at once rather than waited for, so the dialog does not
  // hang around for the round trip. What the server says wins the moment it
  // says anything: a new props object is a pack it just sent, whether or not
  // the value in it changed, and a Go-side Open() right after a manual close
  // has to reopen the dialog.
  const [opened, setOpened] = useState(node.props.opened)
  const [sentProps, setSentProps] = useState(node.props)
  if (sentProps !== node.props) {
    setSentProps(node.props)
    setOpened(node.props.opened)
  }

  const id = node.props.id
  const dismissible = node.props.dismissible

  const close = () => {
    setOpened(false)
    update({ type: "input", id, value: false })
  }

  // Read through a ref by the key handler below, so that closing does not go
  // in its dependencies: re-running the effect on every run would push this
  // dialog back to the top of the stack while another one sits over it.
  const closeRef = useRef(close)
  closeRef.current = close

  // Every open dialog joins the stack, dismissible or not: an undismissible
  // one on top has to swallow the ESC rather than let the one under it take
  // it.
  useEffect(() => {
    if (!opened) {
      return
    }

    openDialogs.push(id)
    return () => {
      const at = openDialogs.lastIndexOf(id)
      if (at !== -1) {
        openDialogs.splice(at, 1)
      }
    }
  }, [opened, id])

  useEffect(() => {
    if (!opened || !dismissible) {
      return
    }

    const onKeyDown = (event: KeyboardEvent) => {
      if (event.key !== "Escape" || event.isComposing) {
        return
      }

      if (openDialogs[openDialogs.length - 1] !== id) {
        return
      }

      if (handledInside(event.target)) {
        return
      }

      closeRef.current()
    }

    window.addEventListener("keydown", onKeyDown, true)
    return () => { window.removeEventListener("keydown", onKeyDown, true) }
  }, [opened, dismissible, id])

  return (
    <Modal
      id={id || undefined}
      title={node.props.title}
      size={SIZES[node.props.width] ?? SIZES.small}
      opened={opened}
      withCloseButton={dismissible}
      closeButtonProps={{ "aria-label": "Close dialog" }}
      // Handled above instead, so that one press closes one dialog.
      closeOnEscape={false}
      closeOnClickOutside={dismissible}
      onClose={close}>
      {node.children.map(child =>
        <TComponent key={child.reactKey} node={child}
          update={update}
          upload={upload}
          theme={theme} />
      )}
    </Modal>
  )
}
