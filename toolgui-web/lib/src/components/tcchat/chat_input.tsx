import React, { useEffect, useState } from "react"
import { createPortal } from "react-dom"
import { ActionIcon, Textarea } from "@mantine/core"
import { IconSend } from "@tabler/icons-react"

import { Props } from "../component_interface"
import { PAGE_BOTTOM_ID } from "../../app/PageBottom"

import '@toolgui-web/lib/src/assets/css/chat.css'

export function TChatInput({ node, update }: Props) {
  const props = node.props
  const [value, setValue] = useState('')

  // The slot is in the same commit as this, so it is looked up once mounted.
  const [slot, setSlot] = useState<HTMLElement | null>(null)
  useEffect(() => {
    if (props.pinned) {
      setSlot(document.getElementById(PAGE_BOTTOM_ID))
    }
  }, [props.pinned])

  const canSend = !props.disabled && value.trim() !== ''
  const send = () => {
    if (!canSend) {
      return
    }

    update({
      type: "form",
      events: [
        { type: "input", id: props.id, value: value },
        { type: "click", id: props.id },
      ],
    })
    setValue('')
  }

  const input = (
    <Textarea id={props.id || undefined}
      className="toolgui-chat-input"
      aria-label={props.placeholder || 'Message'}
      placeholder={props.placeholder}
      maxLength={props.max_length || undefined}
      disabled={props.disabled}
      autosize minRows={1} maxRows={6}
      // Pinned, the slot carries the space below, or it would fall
      // outside the sticky box.
      radius="md" mb={props.pinned ? 0 : "md"}
      value={value}
      onChange={(event) => setValue(event.currentTarget.value)}
      onKeyDown={(event) => {
        // Enter sends and Shift+Enter breaks the line. An Enter that ends
        // an IME composition only picks the characters.
        if (event.key === 'Enter' && !event.shiftKey &&
          !event.nativeEvent.isComposing) {
          event.preventDefault()
          send()
        }
      }}
      rightSectionPointerEvents="all"
      rightSection={
        <ActionIcon variant="subtle" aria-label="Send"
          disabled={!canSend} onClick={send}>
          <IconSend size={18} />
        </ActionIcon>
      } />
  )

  if (props.pinned) {
    return slot ? createPortal(input, slot) : null
  }

  return input
}
