import React from 'react'

import { Props } from '../component_interface'
import { ThemeMode } from '../../util/theme'

import { Prism as SyntaxHighlighter } from 'react-syntax-highlighter'

import { prism, tomorrow } from 'react-syntax-highlighter/dist/esm/styles/prism'

import '@toolgui-web/lib/src/assets/css/code.css'

export interface CodeBlockProps {
  code: string
  // Highlighting language, unset leaves the block unhighlighted.
  lang?: string
  theme: ThemeMode
  id?: string
}

// CodeBlock is the one code block every caller draws, so a fence in a
// Markdown component looks the same as a Code component.
export function CodeBlock({ code, lang, theme, id }: CodeBlockProps) {
  return (
    <SyntaxHighlighter
      id={id}
      className='toolgui-code'
      language={lang}
      style={theme === 'dark' ? tomorrow : prism}>
      {code}
    </SyntaxHighlighter>
  )
}

export function TCode({ node, theme }: Props) {
  return (
    <CodeBlock
      id={node.props.id || undefined}
      code={node.props.code}
      lang={node.props.lang}
      theme={theme} />
  )
}
