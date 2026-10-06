import React, { Suspense, lazy } from "react"
import { Input } from "@mantine/core"

import { Props } from "../component_interface"

// The editor is a chunk of its own, so a page without one never loads it.
const CodeEditor = lazy(() => import("./code_editor"))

export function codeInputBoxStyle(height: number): React.CSSProperties {
  return {
    fontSize: "var(--mantine-font-size-sm)",
    // Lines of 1.5em, plus the content's padding and the border.
    height: `calc(${height} * 1.5em + 10px)`,
  }
}

// A new reset_key remounts the input, which drops the typed value.
export function TCodeInput(props: Props) {
  // Holds the editor's place while it loads, so the page does not jump.
  const fallback = (
    <Input.Wrapper label={props.node.props.label} labelElement="div" mb="md">
      <div style={codeInputBoxStyle(props.node.props.height)} />
    </Input.Wrapper>
  )

  return (
    <Suspense fallback={fallback}>
      <CodeEditor key={props.node.props.reset_key ?? ''} {...props} />
    </Suspense>
  )
}
