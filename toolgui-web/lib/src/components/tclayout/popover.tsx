import React, { useState } from "react"
import { Button, Popover } from "@mantine/core"

import { Props } from "../component_interface"
import { TComponent } from "../factory"

export function TPopover({ node, update, upload, theme }: Props) {
  // The client owns whether this is open: the server only says what the
  // dropdown holds, so a rerun the dropdown itself triggered leaves it open.
  const [opened, setOpened] = useState(false)

  return (
    // The id goes on the Popover, not on the button below: Popover.Target
    // overwrites its child's id with the one it derives from this, which is
    // the id itself when it is given one.
    <Popover id={node.props.id || undefined}
      opened={opened} onChange={setOpened}
      position="bottom-start" shadow="md" withinPortal
      // The dropdown holds whatever the server wrote into it, open or not:
      // nothing here is built lazily, so a widget inside keeps what it holds
      // while the popover is closed.
      keepMounted keepMountedMode="display-none">
      <Popover.Target>
        <Button variant="default"
          disabled={node.props.disabled}
          onClick={() => setOpened(o => !o)}>
          {node.props.label}
        </Button>
      </Popover.Target>

      <Popover.Dropdown className="toolgui-popover-dropdown">
        {
          node.children.map(child =>
            <TComponent key={child.reactKey} node={child}
              update={update}
              upload={upload}
              theme={theme} />
          )
        }
      </Popover.Dropdown>
    </Popover>
  )
}
