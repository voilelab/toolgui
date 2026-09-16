import React, { useMemo, useState } from "react"
import {
  Box, Center, Checkbox, Group, Pagination, Table, Text, TextInput,
  UnstyledButton,
} from "@mantine/core"
import { IconChevronDown, IconChevronUp, IconSearch, IconSelector } from "@tabler/icons-react"

import { stateValues } from "../state"
import { Props } from "../component_interface"

// Column is one column as the server settled it: no default is left open.
interface Column {
  type: "text" | "number" | "datetime"
  align: "left" | "center" | "right"
  width: string
  hidden: boolean
}

// Row is a row carried with the index the page function wrote it at. Filtering
// and sorting reorder the rows, so the index has to travel with them: it is
// what a selection is remembered by and what goes back to the server.
interface Row {
  cells: string[]
  index: number
}

interface Sort {
  column: number
  desc: boolean
}

type Selection = "none" | "single" | "multi"

// Saved is the selection as it is remembered between runs, and is what goes
// back to the server. Both shapes travel together: the keys are what a keyed
// table is read back by, the indices what an unkeyed one is, so a table that
// gains or loses a row key has the other already there to fall back on.
interface Saved {
  indices: number[]
  keys: string[]
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

// compare orders two cells of the same column. dir is 1 ascending and -1
// descending, and is applied to the values only: a cell that does not parse
// sorts last whichever way the column is sorted, which is why descending
// cannot be the ascending order reversed.
function compare(a: string, b: string, type: Column["type"], dir: number): number {
  const ka = sortKey(a, type)
  const kb = sortKey(b, type)

  if (ka === null || kb === null) {
    return ka === kb ? 0 : (ka === null ? 1 : -1)
  }

  if (typeof ka === "string" && typeof kb === "string") {
    return dir * ka.localeCompare(kb)
  }

  return dir * ((ka as number) - (kb as number))
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

export function TDataFrame({ node, update }: Props) {
  const head: string[] = node.props.head
  const columns: Column[] = node.props.columns
  const sortable: boolean = node.props.sortable
  const searchable: boolean = node.props.searchable
  const pageSize: number = node.props.page_size
  const height: string = node.props.height
  const selection: Selection = node.props.selection || "none"

  // Sorting, searching and paging are all local: none of them calls update,
  // so none of them reruns the page function on the server. Picking a row is
  // the one interaction that does.
  const [query, setQuery] = useState("")
  const [sort, setSort] = useState<Sort | null>(null)
  const [page, setPage] = useState(1)

  // The column the rows are named by, null when a selection is a position.
  const rowKey: number | null = node.props.row_key ?? null

  // Nothing is picked until the table is first touched, which is when the
  // default stands in — the same rule Go applies to the state. Kept in
  // stateValues so the pick survives the re-render the server answer brings.
  const [saved, setSaved] = useState<Saved>(
    stateValues[node.props.id]
    || { indices: node.props.default_selection || [], keys: [] })
  const pickable = selection !== "none"

  const rows: Row[] = useMemo(
    () => (node.props.rows || []).map((cells: string[], index: number) =>
      ({ cells, index })), [node.props.rows])

  // What is drawn as picked is the selection resolved against the rows there
  // are now, with the current mode applied, which is what the Go side hands
  // the page function. Both have to be applied here rather than at the point
  // the pick was made: the rows and the mode can change between runs while
  // this component stays mounted, and the hook state does not run again.
  const picked = useMemo(() => {
    if (!pickable) {
      return new Set<number>()
    }

    let idxes: number[]
    if (rowKey !== null && saved.keys.length !== 0) {
      // A name no longer in the table is dropped, rather than leaving its
      // position picked for whatever row moved into it.
      const at = new Map(rows.map(row => [row.cells[rowKey], row.index]))
      idxes = saved.keys
        .map(key => at.get(key))
        .filter((idx): idx is number => idx !== undefined)
        .sort((a, b) => a - b)
    } else {
      idxes = saved.indices.filter(idx => idx >= 0 && idx < rows.length)
    }

    return new Set(selection === "single" ? idxes.slice(0, 1) : idxes)
  }, [saved, selection, pickable, rows, rowKey])

  const shown = useMemo(
    () => head.map((_, i) => i).filter(i => !columns[i].hidden), [head, columns])

  // A hidden column is still searched, so a row can be found by a value it
  // does not show.
  const found = useMemo(() => {
    const needle = query.trim().toLowerCase()
    if (needle === "") {
      return rows
    }

    return rows.filter(row =>
      row.cells.some(cell => cell.toLowerCase().includes(needle)))
  }, [rows, query])

  const sorted = useMemo(() => {
    if (sort === null) {
      return found
    }

    const type = columns[sort.column].type
    const dir = sort.desc ? -1 : 1
    const out = found.slice()
    out.sort((a, b) =>
      compare(a.cells[sort.column], b.cells[sort.column], type, dir))

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

  // commit is the only thing here that talks to the server. The indices are
  // sorted so the page function reads them in row order whatever order they
  // were picked in, which is what the Go side hands back. The names go with
  // them, read off the rows as they are now, which is the only place they can
  // be read: by the next run these indices may mean other rows.
  const commit = (indices: number[]) => {
    const next = [...new Set(indices)].sort((a, b) => a - b)
    const saving: Saved = {
      indices: next,
      keys: rowKey === null ? []
        : next.map(idx => rows[idx].cells[rowKey]),
    }

    stateValues[node.props.id] = saving
    setSaved(saving)
    update({ type: "custom", id: node.props.id, value: saving })
  }

  const toggleRow = (index: number) => {
    if (selection === "single") {
      // Clicking the picked row again clears it, so a single-select table can
      // be emptied without a modifier key.
      commit(picked.has(index) ? [] : [index])
      return
    }

    // Built from picked rather than from what was saved, so a pick that the
    // rows or the mode have since dropped is not carried back in.
    commit(picked.has(index)
      ? [...picked].filter(i => i !== index)
      : [...picked, index])
  }

  // In single mode the row is the only control there is, so it has to be
  // reachable and operable from the keyboard. In multi mode the checkbox
  // already is both, and a focusable row would only add a second tab stop
  // per row without adding anything to do from it.
  const rowKeys = (index: number): React.HTMLAttributes<HTMLTableRowElement> =>
    selection !== "single" ? {} : {
      tabIndex: 0,
      onKeyDown: e => {
        if (e.key !== "Enter" && e.key !== " ") {
          return
        }

        // Space would scroll the page, and Enter would submit the form the
        // table may sit in.
        e.preventDefault()
        toggleRow(index)
      },
    }

  // The head checkbox covers every row the search kept, not just the page on
  // screen: paging is how a long table is read, not how it is divided up.
  const allPicked = sorted.length > 0 && sorted.every(row => picked.has(row.index))
  const somePicked = sorted.some(row => picked.has(row.index))

  const toggleAll = () => {
    const inView = sorted.map(row => row.index)
    if (allPicked) {
      const drop = new Set(inView)
      commit([...picked].filter(i => !drop.has(i)))
      return
    }

    commit([...picked, ...inView])
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
              {selection === "multi" &&
                <Table.Th w="2.5rem">
                  <Checkbox aria-label="select every row"
                    checked={allPicked}
                    indeterminate={somePicked && !allPicked}
                    onChange={toggleAll} />
                </Table.Th>}
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
            {visible.map(row =>
              <Table.Tr key={row.index}
                {...rowKeys(row.index)}
                aria-selected={pickable ? picked.has(row.index) : undefined}
                bg={picked.has(row.index)
                  ? "var(--mantine-color-blue-light)" : undefined}
                style={pickable ? { cursor: "pointer" } : undefined}
                onClick={pickable ? () => toggleRow(row.index) : undefined}>
                {selection === "multi" &&
                  // The click is stopped here so it does not also reach the
                  // row, which would toggle the pick straight back.
                  <Table.Td onClick={e => e.stopPropagation()}>
                    <Checkbox aria-label={`select row ${row.index + 1}`}
                      checked={picked.has(row.index)}
                      onChange={() => toggleRow(row.index)} />
                  </Table.Td>}
                {shown.map(j =>
                  <Table.Td key={j} ta={columns[j].align}>{row.cells[j]}</Table.Td>)}
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
