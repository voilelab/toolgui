import React, { useState } from "react";
import { FileInput } from "@mantine/core"

import { stateGeneration, stateValues } from "../state"
import { Props } from "../component_interface";
import { UploadFunc } from "../../app/Upload";

interface FileMeta {
  name: string
  type: string
  size: number
}

function fileMeta(file: File): FileMeta {
  return { name: file.name, type: file.type, size: file.size }
}

// MAX_FILES is tgframe.MaxFileKeyIndex: the most files one pick may hold.
const MAX_FILES = 1000

// picks is each multi-file component's latest pick and the upload in flight,
// so a new pick stops the old one and waits for its last upload to land.
const picks: Record<string, { gen: number, pending: Promise<unknown> }> = {}

// uploadPick uploads a multi-file pick and sends its metadata. It stops when a
// newer pick or a cleared session (stateGeneration) supersedes it.
export async function uploadPick(
  id: string,
  files: File[],
  upload: UploadFunc,
  send: (value: FileMeta[]) => void,
  setError: (err: string | null) => void,
) {
  if (files.length === 0) {
    return
  }

  if (files.length > MAX_FILES) {
    setError(`Pick at most ${MAX_FILES} files`)
    return
  }
  setError(null)

  const pick = picks[id] ??= { gen: 0, pending: Promise.resolve() }
  const gen = ++pick.gen
  const session = stateGeneration
  const stale = () => pick.gen !== gen || stateGeneration !== session

  // Clear first, so no rerun reads the old pick's names over new bytes.
  send([])
  await pick.pending

  // The i-th file goes under `${id}/${i}`, the key Go reads it from.
  for (let i = 0; i < files.length; i++) {
    if (stale()) {
      return
    }

    const p = upload(files[i], `${id}/${i}`)
    pick.pending = p.catch(() => { })

    const val = await p
    if (stale()) {
      return
    }
    if (!val.ok) {
      console.error(val)
      setError('Upload failed')
      return
    }
  }

  send(files.map(fileMeta))
}

export function TFileupload({ node, update, upload }: Props) {
  const id: string = node.props.id
  const multiple: boolean = !!node.props.multiple
  const [error, setError] = useState<string | null>(null)

  const send = (value: FileMeta | FileMeta[]) => {
    stateValues[id] = value
    update({
      type: "input",
      id: id,
      value: value,
    })
  }

  const handleFileChange = async (picked: File | File[] | null) => {
    if (multiple) {
      await uploadPick(id, (picked as File[] | null) ?? [], upload, send, setError)
      return
    }

    const file = picked as File | null
    if (!file) {
      return
    }

    setError(null)
    const session = stateGeneration
    const val = await upload(file, id)
    if (stateGeneration !== session) {
      return
    }
    if (!val.ok) {
      console.error(val)
      setError('Upload failed')
      return
    }

    send(fileMeta(file))
  };

  const value = stateValues[id]
  let placeholder = 'No file uploaded'
  if (Array.isArray(value)) {
    if (value.length > 0) {
      placeholder = value.map((f: FileMeta) => f.name).join(', ')
    }
  } else if (value) {
    placeholder = value.name
  }

  return (
    <FileInput
      id={id}
      name={id}
      label={node.props.label}
      accept={node.props.accept}
      disabled={node.props.disabled}
      multiple={multiple}
      error={error}
      placeholder={placeholder}
      mb="md"
      onChange={handleFileChange}
    />
  )
}
