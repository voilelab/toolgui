import React from "react"

import { Props } from "../component_interface"

export function TTable({ node }: Props) {
  const head: string[] = node.props.head
  const table: string[][] = node.props.table

  return (
    <div id={node.props.id || undefined} className="table-container">
      <table className="table is-hoverable">
        <thead> <tr> {head.map(s => <th>{s}</th>)} </tr> </thead>
        <tbody>
          {table.map(row => <tr>{row.map(v => <td>{v}</td>)}</tr>)}
        </tbody>
      </table>
    </div>
  )
}
