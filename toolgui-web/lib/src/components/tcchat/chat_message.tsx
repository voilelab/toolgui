import React from "react"
import { Avatar, Box, Group, Paper } from "@mantine/core"
import { IconRobot, IconUser } from "@tabler/icons-react"

import { Props } from "../component_interface"
import { TComponent } from "../factory"

import '@toolgui-web/lib/src/assets/css/chat.css'

const KIND_COLOR: Record<string, string> = {
  user: 'blue',
  assistant: 'orange',
  other: 'gray',
}

function avatarContent(props: any) {
  if (props.avatar_src) {
    return undefined
  }

  if (props.avatar_text) {
    return props.avatar_text
  }

  switch (props.kind) {
    case 'user':
      return <IconUser size={20} />
    case 'assistant':
      return <IconRobot size={20} />
  }
}

export function TChatMessage({ node, update, upload, download, theme }: Props) {
  const props = node.props

  return (
    <Paper id={props.id || undefined}
      className={`toolgui-chat-message toolgui-chat-message-${props.kind}`}
      aria-label={`${props.role} message`}
      // The user's own messages are set apart, as a chat does.
      bg={props.kind === 'user' ? 'var(--mantine-color-gray-light)' : undefined}
      radius="md" p="sm" mb="md">
      <Group wrap="nowrap" align="flex-start" gap="sm">
        <Avatar src={props.avatar_src || undefined}
          alt={props.role}
          color={KIND_COLOR[props.kind] || 'gray'}
          radius="sm">
          {avatarContent(props)}
        </Avatar>
        <Box className="toolgui-chat-message-body">
          {node.children.map(child =>
            <TComponent key={child.reactKey} node={child}
              update={update}
              upload={upload}
              download={download}
              theme={theme} />
          )}
        </Box>
      </Group>
    </Paper>
  )
}
