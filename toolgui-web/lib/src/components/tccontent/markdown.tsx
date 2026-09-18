import React from 'react'

import Markdown from 'react-markdown'
import remarkGfm from 'remark-gfm'
import { Table, Typography } from '@mantine/core'

import { Props } from '../component_interface'
import { remarkEmoji } from '../../util/remark_emoji'
import { CodeBlock } from './code'

import '@toolgui-web/lib/src/assets/css/markdown.css'

export function TMarkdown({ node, theme }: Props) {
  return (
    <Typography className='toolgui-markdown' id={node.props.id || undefined}>
      <Markdown children={node.props.text}
        remarkPlugins={[remarkGfm, remarkEmoji]}
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
          },
          // A GFM table is drawn as the Table component's own table, down to
          // the scroll container that keeps a wide one from widening the page.
          table(props) {
            return (
              <Table.ScrollContainer minWidth={0} type='native'>
                <Table className='toolgui-markdown-table' highlightOnHover>
                  {props.children}
                </Table>
              </Table.ScrollContainer>
            )
          },
          thead(props) {
            return <Table.Thead>{props.children}</Table.Thead>
          },
          tbody(props) {
            return <Table.Tbody>{props.children}</Table.Tbody>
          },
          tr(props) {
            return <Table.Tr>{props.children}</Table.Tr>
          },
          // A cell keeps its own props: a column's GFM alignment arrives as
          // one of them.
          th(props) {
            const { children, node, ...rest } = props
            return <Table.Th {...rest}>{children}</Table.Th>
          },
          td(props) {
            const { children, node, ...rest } = props
            return <Table.Td {...rest}>{children}</Table.Td>
          }
        }}
      />
    </Typography>
  )
}
