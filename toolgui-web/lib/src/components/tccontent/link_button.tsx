import React from "react"
import { Button } from "@mantine/core"

import { Props } from "../component_interface"
import { mantineColor } from "../../util/color"
import { emojize } from "../../util/emoji"

export function TLinkButton({ node }: Props) {
  const color = mantineColor(node.props.color)

  // An anchor, not a button: it navigates, so it has to open in a new tab on
  // a middle click and offer a copyable url the way any link does.
  return (
    <Button id={node.props.id || undefined}
      component="a"
      href={node.props.url}
      color={color}
      variant={color ? 'filled' : 'default'}>
      {emojize(node.props.text)}
    </Button>
  )
}
