import React from "react";
import { FileInput } from "@mantine/core"

import { stateValues } from "../state"
import { Props } from "../component_interface";

interface FileMeta {
  name: string
  type: string
  size: number
}

function fileMeta(file: File): FileMeta {
  return { name: file.name, type: file.type, size: file.size }
}

export function TFileupload({ node, update, upload }: Props) {
  const id: string = node.props.id
  const multiple: boolean = !!node.props.multiple

  const handleFileChange = async (picked: File | File[] | null) => {
    let value: FileMeta | FileMeta[]

    if (multiple) {
      const files = (picked as File[] | null) ?? []
      if (files.length === 0) {
        return
      }

      // The i-th file goes under `${id}/${i}`, the key Go reads it from.
      for (let i = 0; i < files.length; i++) {
        const val = await upload(files[i], `${id}/${i}`)
        if (!val.ok) {
          console.error(val)
          return
        }
      }

      value = files.map(fileMeta)
    } else {
      const file = picked as File | null
      if (!file) {
        return
      }

      const val = await upload(file, id)
      if (!val.ok) {
        console.error(val)
        return
      }

      value = fileMeta(file)
    }

    stateValues[id] = value
    update({
      type: "input",
      id: id,
      value: value,
    })
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
      placeholder={placeholder}
      mb="md"
      onChange={handleFileChange}
    />
  )
}
