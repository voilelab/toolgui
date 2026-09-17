import React from "react"
import { Button } from "@mantine/core"

import { Props } from '../component_interface'
import { mantineColor } from "../../util/color"

// How long the blob URL is left alive after the click. The browser resolves it
// while the click is dispatched and holds what it needs from there on, but not
// all of them have let go by the time the handler returns -- and revoking too
// early is a download that never starts. The blob is backed by whatever the
// transport read the file out of, so waiting costs no heap.
const REVOKE_DELAY_MS = 10000

// TDownloadFile saves a file the page offered, fetched by the token the
// component carries rather than out of the component itself.
//
// The href is a blob URL made here, so there is no URI for Go to put in front
// of the DOM. Saving one still wants an anchor: a click is the only thing that
// reaches the browser's download path, and React has nothing for it.
export function TDownloadFile({ node, update, download }: Props) {
  const color = mantineColor(node.props.color)

  // The click is reported after the bytes are in hand, not before. A rerun is
  // what the report causes, and a rerun may offer a different file -- which
  // retires this token -- so fetching first is what keeps the click from
  // cancelling its own download.
  const save = async () => {
    const res = await download(node.props.token)
    if (!res.ok || !res.blob) {
      console.error('download', res.error)
      return
    }

    const url = URL.createObjectURL(res.blob)

    const link = document.createElement('a')
    link.setAttribute('download', node.props.filename)
    link.href = url
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)

    setTimeout(() => { URL.revokeObjectURL(url) }, REVOKE_DELAY_MS)

    update({
      type: "click",
      id: node.props.id,
    })
  }

  return (
    <Button id={node.props.id}
      color={color}
      variant={color ? 'filled' : 'default'}
      disabled={node.props.disabled}
      onClick={() => { save().catch((e) => { console.error(e) }) }}>
      {node.props.text}
    </Button>
  )
}
