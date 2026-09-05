import React from "react";

import { stateValues } from "../state"
import { Props } from "../component_interface";

export function TFileupload({ node, update, upload }: Props) {
  const handleFileChange: React.ChangeEventHandler<HTMLInputElement> = async (e) => {
    e.preventDefault();

    if (!e.target.files) {
      return
    }

    const file = e.target.files[0]

    upload(file).then(val => {
      if (!val.ok) {
        console.error(val)
        return
      }
      const newFile = {
        name: file.name,
        type: file.type,
        size: file.size,
      }

      stateValues[e.target.id] = newFile
      update({
        type: "input",
        id: e.target.id,
        value: newFile,
      })
    })
  };

  const file = stateValues[node.props.id]

  return (
    <div className="field">
      <label className="label">{node.props.label}</label>
      <div className="file has-name">
        <label className="file-label">
          <input className="file-input" type="file"
            id={node.props.id}
            name={node.props.id}
            accept={node.props.accept}
            onChange={handleFileChange} />
          <span className="file-cta">
            <span className="file-icon">
              <i className="fas fa-upload"></i>
            </span>
            <span className="file-label"> {node.props.label} </span>
          </span>
          <span className="file-name"> {file ? file.name : 'No file uploaded'} </span>
        </label>
      </div>
    </div>
  )
}
