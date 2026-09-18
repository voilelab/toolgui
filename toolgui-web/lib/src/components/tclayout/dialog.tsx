import React, { useEffect, useState } from "react"
import { Modal, getDefaultZIndex } from "@mantine/core"

import { Props } from "../component_interface"
import { TComponent } from "../factory"
import { useEscapeToClose, useOverlay } from "./overlay_stack"

// Classes of our own on the parts of the dialog the sticky offset below is
// measured from, so none of Mantine's internal ones are reached for.
const HEADER_CLASS = "toolgui-dialog-header"
const CONTENT_CLASS = "toolgui-dialog-content"

// Mantine's size for each of the widths the server names.
const SIZES: { [width: string]: string } = {
  small: "sm",
  medium: "md",
  large: "lg",
}

export function TDialog({ node, update, upload, download, theme }: Props) {
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

  // Mantine scrolls the dialog's content and keeps the header sticky at the
  // top of it, so anything else sticky in the body -- a Toolbar -- has to
  // start below that header rather than under it. The height is published on
  // the scroller as --tg-sticky-top for those rules to read, and watched
  // because a title that wraps changes it.
  //
  // The body is found through a node of its own rather than the Modal's ref,
  // which lands on the root: the content is mounted by the open transition,
  // so a ref taken when `opened` changes is still empty. Kept in state so
  // that mounting it is what runs the effect.
  const [bodyEl, setBodyEl] = useState<HTMLElement | null>(null)
  useEffect(() => {
    const content = bodyEl?.closest<HTMLElement>("." + CONTENT_CLASS)
    const header = content?.querySelector<HTMLElement>("." + HEADER_CLASS)
    if (!content || !header) {
      return
    }

    const sync = () => content.style.setProperty(
      "--tg-sticky-top", `${header.getBoundingClientRect().height}px`)

    sync()

    const observer = new ResizeObserver(sync)
    observer.observe(header)
    return () => observer.disconnect()
  }, [bodyEl])

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
      classNames={{ header: HEADER_CLASS, content: CONTENT_CLASS }}
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
      {/* display:contents, so marking the body costs it no box of its own. */}
      <div ref={setBodyEl} style={{ display: "contents" }}>
        {node.children.map(child =>
          <TComponent key={child.reactKey} node={child}
            update={update}
            upload={upload}
            download={download}
            theme={theme} />
        )}
      </div>
    </Modal>
  )
}
