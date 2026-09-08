import React from "react";
import { FileInput } from "@mantine/core"

import { stateValues } from "../state"
import { Props } from "../component_interface";

export function TFileupload({ node, update, upload }: Props) {
  const handleFileChange = async (file: File | null) => {
    if (!file) {
      return
    }

    const val = await upload(file, node.props.id)
    if (!val.ok) {
      console.error(val)
      return
    }

    const newFile = {
      name: file.name,
      type: file.type,
      size: file.size,
    }

    stateValues[node.props.id] = newFile
    update({
      type: "input",
      id: node.props.id,
      value: newFile,
    })
  };

  const file = stateValues[node.props.id]

  return (
    <FileInput
      id={node.props.id}
      name={node.props.id}
      label={node.props.label}
      accept={node.props.accept}
      disabled={node.props.disabled}
      placeholder={file ? file.name : 'No file uploaded'}
      mb="md"
      onChange={handleFileChange}
    />
  )
}
