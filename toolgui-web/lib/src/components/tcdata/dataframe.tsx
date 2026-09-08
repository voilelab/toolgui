import React, { useMemo, useState } from "react"
import {
  Box, Center, Group, Pagination, Table, Text, TextInput, UnstyledButton,
} from "@mantine/core"
import { IconChevronDown, IconChevronUp, IconSearch, IconSelector } from "@tabler/icons-react"

import { Props } from "../component_interface"

// Column is one column as the server settled it: no default is left open.
interface Column {
  type: "text" | "number" | "datetime"
  align: "left" | "center" | "right"
  width: string
  hidden: boolean
}

interface Sort {
  column: number
  desc: boolean
}

// sortKey reads a cell as the value its column type says it holds. Cells that
// do not parse come back as null and are kept together at the end, so a
// stray "-" never lands in the middle of the numbers.
function sortKey(value: string, type: Column["type"]): number | string | null {
  switch (type) {
    case "number": {
      // Number("") is 0, which would sort an empty cell among the values.
      const trimmed = value.trim()
      if (trimmed === "") {
        return null
      }

      const n = Number(trimmed)
      return Number.isNaN(n) ? null : n
    }
    case "datetime": {
      const t = Date.parse(value)
      return Number.isNaN(t) ? null : t
    }
    default:
      return value
  }
}

function compare(a: string, b: string, type: Column["type"]): number {
  const ka = sortKey(a, type)
  const kb = sortKey(b, type)

  if (ka === null || kb === null) {
    return ka === kb ? 0 : (ka === null ? 1 : -1)
  }

  if (typeof ka === "string" && typeof kb === "string") {
    return ka.localeCompare(kb)
  }

  return (ka as number) - (kb as number)
}

// SortButton is a column head that toggles through ascending, descending and
// back to the order the rows arrived in.
function SortButton({ label, sort, onSort }: {
  label: string
  sort: "asc" | "desc" | null
  onSort: () => void
}) {
  const Icon = sort === "asc" ? IconChevronUp
    : sort === "desc" ? IconChevronDown : IconSelector

  return (
    <UnstyledButton onClick={onSort} style={{ width: "100%" }}
      aria-label={`sort by ${label}`}>
      <Group justify="space-between" wrap="nowrap" gap="xs">
        <Text fw={700} fz="sm">{label}</Text>
        <Icon size={14} stroke={1.5} />
      </Group>
    </UnstyledButton>
  )
}

export function TDataFrame({ node }: Props) {
  const head: string[] = node.props.head
  const rows: string[][] = node.props.rows || []
  const columns: Column[] = node.props.columns
  const sortable: boolean = node.props.sortable
  const searchable: boolean = node.props.searchable
  const pageSize: number = node.props.page_size
  const height: string = node.props.height

  // Sorting, searching and paging are all local: none of them calls update,
  // so none of them reruns the page function on the server.
  const [query, setQuery] = useState("")
  const [sort, setSort] = useState<Sort | null>(null)
  const [page, setPage] = useState(1)

  const shown = useMemo(
    () => head.map((_, i) => i).filter(i => !columns[i].hidden), [head, columns])

  // A hidden column is still searched, so a row can be found by a value it
  // does not show.
  const found = useMemo(() => {
    const needle = query.trim().toLowerCase()
    if (needle === "") {
      return rows
    }

    return rows.filter(row => row.some(cell => cell.toLowerCase().includes(needle)))
  }, [rows, query])

  const sorted = useMemo(() => {
    if (sort === null) {
      return found
    }

    const type = columns[sort.column].type
    const out = found.slice()
    out.sort((a, b) => compare(a[sort.column], b[sort.column], type))
    if (sort.desc) {
      out.reverse()
    }

    return out
  }, [found, sort, columns])

  // Only one page is ever in the DOM, which is what keeps a table of tens of
  // thousands of rows smooth.
  const pageCount = Math.max(1, Math.ceil(sorted.length / pageSize))
  const current = Math.min(page, pageCount)
  const start = (current - 1) * pageSize
  const visible = sorted.slice(start, start + pageSize)

  const toggleSort = (column: number) => {
    setPage(1)
    setSort(prev => {
      if (prev === null || prev.column !== column) {
        return { column, desc: false }
      }

      return prev.desc ? null : { column, desc: true }
    })
  }

  return (
    <Box id={node.props.id || undefined}>
      {searchable &&
        <TextInput mb="xs" value={query}
          placeholder="Search"
          aria-label="search the table"
          leftSection={<IconSearch size={16} stroke={1.5} />}
          onChange={e => { setQuery(e.currentTarget.value); setPage(1) }} />}

      {/* minWidth 0: the table is never forced wider than the page, it only
          scrolls when its own content overflows. */}
      <Table.ScrollContainer minWidth={0} type="native"
        maxHeight={height || undefined}>
        <Table highlightOnHover stickyHeader={height !== ""}>
          <Table.Thead>
            <Table.Tr>
              {shown.map(i =>
                <Table.Th key={i} w={columns[i].width || undefined}
                  ta={columns[i].align}>
                  {sortable
                    ? <SortButton label={head[i]}
                      sort={sort?.column === i ? (sort.desc ? "desc" : "asc") : null}
                      onSort={() => toggleSort(i)} />
                    : head[i]}
                </Table.Th>)}
            </Table.Tr>
          </Table.Thead>
          <Table.Tbody>
            {visible.map((row, i) =>
              <Table.Tr key={start + i}>
                {shown.map(j =>
                  <Table.Td key={j} ta={columns[j].align}>{row[j]}</Table.Td>)}
              </Table.Tr>
            )}
          </Table.Tbody>
        </Table>
      </Table.ScrollContainer>

      {sorted.length === 0 &&
        <Text c="dimmed" fz="sm" ta="center" py="md">No rows</Text>}

      {pageCount > 1 &&
        <Center mt="xs">
          <Pagination total={pageCount} value={current} onChange={setPage}
            size="sm" withEdges />
        </Center>}
    </Box>
  )
}
