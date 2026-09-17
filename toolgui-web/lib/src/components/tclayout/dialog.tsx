import React, { useState } from "react"
import { Modal } from "@mantine/core"

import { Props } from "../component_interface"
import { TComponent } from "../factory"

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

  const dismissible = node.props.dismissible

  return (
    <Modal
      id={node.props.id || undefined}
      title={node.props.title}
      size={SIZES[node.props.width] ?? SIZES.small}
      opened={opened}
      withCloseButton={dismissible}
      closeButtonProps={{ "aria-label": "Close dialog" }}
      closeOnEscape={dismissible}
      closeOnClickOutside={dismissible}
      onClose={() => {
        setOpened(false)
        update({
          type: "input",
          id: node.props.id,
          value: false,
        })
      }}>
      {node.children.map(child =>
        <TComponent key={child.reactKey} node={child}
          update={update}
          upload={upload}
          theme={theme} />
      )}
    </Modal>
  )
}
