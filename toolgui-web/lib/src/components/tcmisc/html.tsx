import React from "react"

import DOMPurify from "dompurify"

import { Props } from "../component_interface";

export function THtml({ node }: Props) {
  return (
    <div id={node.props.id || undefined} dangerouslySetInnerHTML={{ __html: DOMPurify.sanitize(node.props.html) }} />
  )
}
