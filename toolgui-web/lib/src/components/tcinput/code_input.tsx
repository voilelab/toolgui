import React, { Suspense, lazy } from "react"
import { Input } from "@mantine/core"

import { Props } from "../component_interface"

// The editor is a chunk of its own, so a page without one never loads it.
const CodeEditor = lazy(() => import("./code_editor"))

// lines is n lines of 1.5em, plus the content's 4px padding top and bottom.
const lines = (n: number) => `calc(${n} * 1.5em + 8px)`

// codeInputBoxStyle sizes the editor's box: it grows with the code from
// height lines, up to maxHeight lines when that is set. The editor reads the
// two bounds from these variables.
export function codeInputBoxStyle(
  height: number, maxHeight?: number): React.CSSProperties {

  return {
    fontSize: "var(--mantine-font-size-sm)",
    ["--toolgui-code-min" as string]: lines(height),
    // Plus the border, as the bound is on the editor's outer box.
    ["--toolgui-code-max" as string]:
      maxHeight ? `calc(${lines(maxHeight)} + 2px)` : "none",
  }
}

// A new reset_key remounts the input, which drops the typed value.
export function TCodeInput(props: Props) {
  // Holds the editor's place while it loads, so the page does not jump.
  const fallback = (
    <Input.Wrapper label={props.node.props.label} labelElement="div" mb="md">
      <div style={{
        ...codeInputBoxStyle(props.node.props.height),
        minHeight: `calc(${lines(props.node.props.height)} + 2px)`,
      }} />
    </Input.Wrapper>
  )

  return (
    <Suspense fallback={fallback}>
      <CodeEditor key={props.node.props.reset_key ?? ''} {...props} />
    </Suspense>
  )
}
