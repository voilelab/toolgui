import React from 'react'

import Markdown from 'react-markdown'
import { Typography } from '@mantine/core'

import { Props } from '../component_interface'
import { remarkEmoji } from '../../util/remark_emoji'
import { CodeBlock } from './code'

export function TMarkdown({ node, theme }: Props) {
  return (
    <Typography id={node.props.id || undefined}>
      <Markdown children={node.props.text}
        remarkPlugins={[remarkEmoji]}
        components={{
          a(props) {
            const { children, className, node, ...rest } = props
            return (
              <a {...rest} target='_blank'>
                {children}
              </a>
            )
          },
          // A code block is drawn by the Code component's own block, and
          // replaces the <pre> rather than sitting in it: the block brings its
          // own <pre>, and Typography styles one of those it finds.
          pre(props) {
            const { children, node, ...rest } = props

            const child = React.Children.toArray(children)[0]
            if (!React.isValidElement(child) || child.type !== 'code') {
              return <pre {...rest}>{children}</pre>
            }

            const codeProps = child.props as {
              className?: string
              children?: React.ReactNode
            }
            const match = /language-([\w-]+)/.exec(codeProps.className || '')

            return (
              <CodeBlock
                code={String(codeProps.children).replace(/\n$/, '')}
                lang={match ? match[1] : undefined}
                theme={theme} />
            )
          }
        }}
      />
    </Typography>
  )
}
