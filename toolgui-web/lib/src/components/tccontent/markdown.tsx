import React from 'react'

import Markdown from 'react-markdown'

import { Prism as SyntaxHighlighter } from 'react-syntax-highlighter'
import { prism, tomorrow } from 'react-syntax-highlighter/dist/esm/styles/prism'

import { Props } from '../component_interface'

export function TMarkdown({ node, theme }: Props) {
  return (
    <div className="content">
      <Markdown children={node.props.text}
        components={{
          a(props) {
            const { children, className, node, ...rest } = props
            return (
              <a {...rest} target='_blank'>
                {children}
              </a>
            )
          },
          code(props) {
            const { children, className, node, ...rest } = props
            const match = /language-(\w+)/.exec(className || '')
            if (!match) {
              return (
                <code {...rest} className={className}>
                  {children}
                </code>
              )
            }

            const lang = match[1]
            const code = String(children).replace(/\n$/, '')

            return (
              <SyntaxHighlighter
                PreTag="div"
                language={lang}
                style={theme === 'dark' ? tomorrow : prism}
              >
                {code}
              </SyntaxHighlighter>
            )
          }
        }}
      />
    </div>
  )
}