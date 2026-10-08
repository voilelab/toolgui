import React from "react"
import { Anchor } from "@mantine/core"

import { Props } from '../component_interface'
import { emojize } from '../../util/emoji'
import { usePageNav } from '../../app/PageNav'

// TPageLink is a real link to a page of this app, so middle-click and "copy
// link" get the page URL. A plain click goes through the app's navigation.
export function TPageLink({ node }: Props) {
  const nav = usePageNav()
  const { page, query } = node.props

  return (
    <Anchor id={node.props.id || undefined} href={nav.href(page, query)}
      onClick={(e: React.MouseEvent<HTMLAnchorElement>) => {
        // A modified click opens a tab or a window: the browser's to handle.
        if (e.button !== 0 || e.metaKey || e.ctrlKey || e.shiftKey || e.altKey) {
          return
        }

        e.preventDefault()
        nav.navigate(page, query)
      }}>
      {emojize(node.props.text)}
    </Anchor>
  )
}
