import React, { useEffect, useRef, useState, useSyncExternalStore } from "react"
import { Modal, getDefaultZIndex } from "@mantine/core"

import { Props } from "../component_interface"
import { TComponent } from "../factory"

// Mantine's size for each of the widths the server names.
const SIZES: { [width: string]: string } = {
  small: "sm",
  medium: "md",
  large: "lg",
}

// openDialogs is every dialog on screen, in the order they opened. It decides
// both which one answers ESC and which one is painted over the others, so the
// two can never disagree.
//
// Both need saying because Mantine settles neither: its own closeOnEscape is a
// window handler per Modal with no notion of a stack, and every Modal takes
// the same z-index, leaving the paint order to the DOM -- which follows where
// the page writes a dialog, not when it opened. Left alone, a dialog opened
// from inside one written after it lands underneath, unclickable, while ESC
// closes it rather than the one on top.
const openDialogs: string[] = []
const listeners = new Set<() => void>()

function announce() {
  listeners.forEach(notify => notify())
}

function subscribeToDialogs(notify: () => void) {
  listeners.add(notify)
  return () => { listeners.delete(notify) }
}

function openedDialog(id: string) {
  openDialogs.push(id)
  announce()
}

function closedDialog(id: string) {
  const at = openDialogs.lastIndexOf(id)
  if (at === -1) {
    return
  }

  openDialogs.splice(at, 1)
  announce()
}

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

    openedDialog(id)
    return () => { closedDialog(id) }
  }, [opened, id])

  const depth = useSyncExternalStore(
    subscribeToDialogs, () => openDialogs.indexOf(id))

  // One step per dialog below this one, so the newest is on top. Kept under
  // the popover elevation, or a Select dropdown opened inside a dialog would
  // fall behind it.
  const zIndex = Math.min(
    getDefaultZIndex("modal") + Math.max(depth, 0),
    getDefaultZIndex("popover") - 1)

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
      zIndex={zIndex}
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
