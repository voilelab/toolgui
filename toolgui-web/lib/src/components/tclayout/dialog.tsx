import React, { useState } from "react"
import { Modal, getDefaultZIndex } from "@mantine/core"

import { Props } from "../component_interface"
import { TComponent } from "../factory"
import { useEscapeToClose, useOverlay } from "./overlay_stack"

// Mantine's size for each of the widths the server names.
const SIZES: { [width: string]: string } = {
  small: "sm",
  medium: "md",
  large: "lg",
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

  // Every open dialog joins the stack, dismissible or not: an undismissible
  // one on top has to swallow the ESC rather than let the one under it take
  // it.
  const { depth, isTop } = useOverlay(id, "dialog", opened)
  useEscapeToClose(opened && dismissible && isTop, close)

  // One step per dialog below this one, so the newest is on top. Kept under
  // the popover elevation, or a Select dropdown opened inside a dialog would
  // fall behind it.
  const zIndex = Math.min(
    getDefaultZIndex("modal") + Math.max(depth, 0),
    getDefaultZIndex("popover") - 1)

  return (
    <Modal
      id={id || undefined}
      title={node.props.title}
      size={SIZES[node.props.width] ?? SIZES.small}
      opened={opened}
      zIndex={zIndex}
      withCloseButton={dismissible}
      closeButtonProps={{ "aria-label": "Close dialog" }}
      // Handled through the overlay stack instead, so that one press closes
      // one overlay.
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
