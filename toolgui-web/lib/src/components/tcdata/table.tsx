import React from "react"
import { Table } from "@mantine/core"

import { Props } from "../component_interface"

export function TTable({ node }: Props) {
  const head: string[] = node.props.head
  const table: string[][] = node.props.table

  // minWidth 0: the table is never forced wider than the page, it only
  // scrolls when its own content overflows.
  return (
    <Table.ScrollContainer id={node.props.id || undefined}
      minWidth={0} type="native">
      <Table highlightOnHover>
        <Table.Thead>
          <Table.Tr>
            {head.map((s, i) => <Table.Th key={i}>{s}</Table.Th>)}
          </Table.Tr>
        </Table.Thead>
        <Table.Tbody>
          {table.map((row, i) =>
            <Table.Tr key={i}>
              {row.map((v, j) => <Table.Td key={j}>{v}</Table.Td>)}
            </Table.Tr>
          )}
        </Table.Tbody>
      </Table>
    </Table.ScrollContainer>
  )
}
