import React, { useEffect, useRef } from "react"

import { Props } from "../component_interface"

import katex from "katex"

import '@toolgui-web/lib/src/assets/css/latex.css'

export function TLatex({ node }: Props) {

  const containerRef = useRef<HTMLDivElement>();

  useEffect(() => {
    katex.render(node.props.latex, containerRef.current, {
      displayMode: false,
      throwOnError: false,
    });
  }, [node.props.latex]);

  return <div id={node.props.id || undefined} ref={containerRef} />
}
