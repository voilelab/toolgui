import React, { useState } from "react";
import { FileInput, Text } from "@mantine/core"

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

// formatSize renders a byte count, e.g. 1536 -> "1.5 KB".
export function formatSize(bytes: number): string {
  const units = ['B', 'KB', 'MB', 'GB', 'TB']
  let n = bytes
  let i = 0
  while (n >= 1024 && i < units.length - 1) {
    n /= 1024
    i++
  }
  return i === 0 ? `${n} B` : `${n.toFixed(1)} ${units[i]}`
}

// acceptLabel renders an accept string, e.g. ".csv,image/png" -> "CSV, image/png".
export function acceptLabel(accept: string | undefined): string {
  if (!accept) {
    return ''
  }
  return accept.split(',')
    .map(s => s.trim())
    .filter(s => s)
    .map(s => s.startsWith('.') ? s.slice(1).toUpperCase() : s)
    .join(', ')
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
  const files: FileMeta[] = Array.isArray(value) ? value : value ? [value] : []
  const placeholder = files.length > 0 ?
    files.map(f => f.name).join(', ') : 'No file uploaded'

  const accept = acceptLabel(node.props.accept)
  const info = [multiple ? 'Multiple files' : 'Single file']
  if (accept) {
    info.push(accept)
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
      inputWrapperOrder={['label', 'input', 'description', 'error']}
      description={
        <Text component="span" size="xs" display="block">
          {info.join(' • ')}
          {files.map((f, i) => (
            <Text component="span" size="xs" display="block" key={i}>
              {f.name} • {formatSize(f.size)}{f.type ? ` • ${f.type}` : ''}
            </Text>
          ))}
        </Text>
      }
    />
  )
}
