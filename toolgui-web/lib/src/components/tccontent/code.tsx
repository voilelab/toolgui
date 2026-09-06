import React from 'react'

import { Props } from '../component_interface'

import { Prism as SyntaxHighlighter } from 'react-syntax-highlighter'

import { prism, tomorrow } from 'react-syntax-highlighter/dist/esm/styles/prism'

export function TCode({ node, theme }: Props) {
  return (
    <SyntaxHighlighter
      id={node.props.id || undefined}
      language={node.props.lang}
      style={theme === 'dark' ? tomorrow : prism}>
      {node.props.code}
    </SyntaxHighlighter>
  )
}
