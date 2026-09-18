import React, { useContext, useState } from "react"
import { Button, Menu } from "@mantine/core"

import { Props } from "../component_interface"
import { mantineColor } from "../../util/color"
import { useOverlay } from "../tclayout/overlay_stack"
import { FormSubmitContext } from "./form_context"

// One entry of the dropdown, as the server writes it.
interface Item {
  id: string
  label: string
}

export function TMenu({ node, update }: Props) {
  const color = mantineColor(node.props.color)

  // Null outside a form. Inside one, picking an item is what sends it: a menu
  // item is an action to take now, the way a button press is, so it cannot sit
  // in the form's queue waiting for something else to send it.
  const submitForm = useContext(FormSubmitContext)

  // The client owns whether this is open, the same way a popover does: the
  // server only says what the dropdown holds, so a rerun an item started
  // leaves the menu where the click left it.
  const [opened, setOpened] = useState(false)

  // An open menu traps focus in its dropdown, so Mantine's own ESC -- which
  // it hangs off the dropdown -- always reaches it, unlike a popover's. What
  // it cannot do is keep a dialog underneath from taking the same press, so
  // the menu joins the overlay stack: on top, it is the dialog that stands
  // down, and the second press closes the dialog.
  useOverlay(node.props.id, "popover", opened)

  return (
    // The id goes on the Menu, not on the button below: Menu.Target overwrites
    // its child's id with the one it derives from this, which is the id itself
    // when it is given one.
    <Menu id={node.props.id || undefined}
      opened={opened} onChange={setOpened}
      position="bottom-start" shadow="md" withinPortal>
      <Menu.Target>
        <Button
          color={color}
          // A colourless button stays neutral, as [TButton] does.
          variant={color ? 'filled' : 'default'}
          disabled={node.props.disabled}>
          {node.props.label}
        </Button>
      </Menu.Target>

      <Menu.Dropdown className="toolgui-menu-dropdown">
        {
          (node.props.items as Item[] ?? []).map(item =>
            // The click carries the item's id, not the menu's: that is what
            // tells the server which item it was.
            <Menu.Item key={item.id} id={item.id}
              onClick={() => {
                update({ type: "click", id: item.id })

                // Inside a form the update above only queued the click. This
                // sends it, with the inputs queued before it, so the run that
                // reports the pick is the one that reads the new values. It
                // also keeps the queue from holding two clicks at once, where
                // the later one would be the only pick the server saw.
                submitForm?.()
              }}>
              {item.label}
            </Menu.Item>
          )
        }
      </Menu.Dropdown>
    </Menu>
  )
}
