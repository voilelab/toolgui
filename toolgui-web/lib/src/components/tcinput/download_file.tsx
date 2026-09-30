import React, { useContext, useEffect, useRef } from "react"
import { Button } from "@mantine/core"

import { Props } from '../component_interface'
import { mantineColor } from "../../util/color"
import { FormSubmitContext } from "./form_context"

// How long the blob URL is left alive after the click. The browser resolves it
// while the click is dispatched and holds what it needs from there on, but not
// all of them have let go by the time the handler returns -- and revoking too
// early is a download that never starts. The blob is backed by whatever the
// transport read the file out of, so waiting costs no heap.
const REVOKE_DELAY_MS = 10000

// saveBlob hands blob to the browser's download path under filename.
function saveBlob(blob: Blob, filename: string) {
  const url = URL.createObjectURL(blob)

  const link = document.createElement('a')
  link.setAttribute('download', filename)
  link.href = url
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)

  setTimeout(() => { URL.revokeObjectURL(url) }, REVOKE_DELAY_MS)
}

// TDownloadFile saves a file the page offered, fetched by the token the
// component carries rather than out of the component itself.
//
// The href is a blob URL made here, so there is no URI for Go to put in front
// of the DOM. Saving one still wants an anchor: a click is the only thing that
// reaches the browser's download path, and React has nothing for it.
export function TDownloadFile(props: Props) {
  return props.node.props.lazy
    ? <TDownloadFileLazy {...props} />
    : <TDownloadFileEager {...props} />
}

// TDownloadFileLazy is DownloadFileFunc: the click goes first and the file is
// made by the run it starts. That run's pack carries a new serial, and a
// serial arriving after this button's own click is the file to save.
function TDownloadFileLazy({ node, update, download }: Props) {
  const color = mantineColor(node.props.color)

  // Set by a click, cleared by the save it asked for. A serial that shows up
  // without one, as on a remount, saves nothing.
  const pending = useRef(false)

  const { serial, token, filename } = node.props

  useEffect(() => {
    if (!serial || !pending.current) {
      return
    }
    pending.current = false

    download(token).then((res) => {
      if (!res.ok || !res.blob) {
        console.error('download', res.error)
        return
      }
      saveBlob(res.blob, filename)
    }).catch((e) => { console.error(e) })
  }, [serial])

  return (
    <Button id={node.props.id}
      color={color}
      variant={color ? 'filled' : 'default'}
      disabled={node.props.disabled}
      onClick={() => {
        pending.current = true
        update({
          type: "click",
          id: node.props.id,
        })
      }}>
      {node.props.text}
    </Button>
  )
}

function TDownloadFileEager({ node, update, download }: Props) {
  const color = mantineColor(node.props.color)

  // Null outside a form. Inside one, an update only queues: nothing reruns the
  // page until the form is submitted.
  const inForm = useContext(FormSubmitContext) !== null

  // When the click is reported depends on what reporting does. On its own it
  // reruns the page, and a rerun that offers a different file retires this
  // token, so the bytes are fetched first and the click follows them. In a
  // form it reruns nothing, so the click is queued at once -- waiting there
  // would strand it behind a submit the user makes while the file is still on
  // its way, and it would only go out with the submit after that.
  const save = async () => {
    // Read before the await, not after. A node keeps its identity across runs
    // and has its props replaced in place, so a rerun that lands while the
    // fetch is out would otherwise save these bytes under the next run's
    // filename, or report the click under whatever id now sits here.
    const { token, filename, id } = node.props

    const report = () => {
      update({
        type: "click",
        id: id,
      })
    }

    if (inForm) {
      report()
    }

    const res = await download(token)
    if (!res.ok || !res.blob) {
      console.error('download', res.error)
      return
    }

    saveBlob(res.blob, filename)

    if (!inForm) {
      report()
    }
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
