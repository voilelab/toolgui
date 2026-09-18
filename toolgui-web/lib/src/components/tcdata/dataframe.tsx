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

// RowKey is what one row is remembered by: the index the page function wrote
// it at, or the key the page function named it with when row_keys is set. A
// key survives the rows changing underneath the table, an index does not.
type RowKey = number | string

// Row is a row carried with both. Filtering and sorting reorder the rows, so
// they have to travel with them: the key is what a selection is remembered by
// and what goes back to the server, the index is where the row sits now.
interface Row {
  cells: string[]
  index: number
  key: RowKey
}

interface Sort {
  column: number
  desc: boolean
}

type Selection = "none" | "single" | "multi"

// noKeys stands in for a server that sends no row_keys at all, kept module
// level so that the memo the rows hang off is not invalidated every render by
// a fresh empty array.
const noKeys: string[] = []

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
  const rowCells: string[][] = node.props.rows
  const columns: Column[] = node.props.columns
  const sortable: boolean = node.props.sortable
  const searchable: boolean = node.props.searchable
  const pageSize: number = node.props.page_size
  const height: string = node.props.height
  const selection: Selection = node.props.selection || "none"

  // Empty is how the server says the table is positional, so the rows are
  // remembered by index and the selection goes back as one.
  const rowKeyList: string[] = node.props.row_keys ?? noKeys
  const keyed = rowKeyList.length !== 0
  const keyOf = (index: number): RowKey => keyed ? rowKeyList[index] : index

  // Sorting, searching and paging are all local: none of them calls update,
  // so none of them reruns the page function on the server. Picking a row is
  // the one interaction that does.
  const [query, setQuery] = useState("")
  const [sort, setSort] = useState<Sort | null>(null)
  const [page, setPage] = useState(1)

  // Nothing is picked until the table is first touched, which is when the
  // default stands in — the same rule Go applies to the state. Kept in
  // stateValues so the pick survives the re-render the server answer brings.
  // The default arrives as positions, so a keyed table reads it through the
  // keys before it can compare it with anything else it holds.
  const [selected, setSelected] = useState<RowKey[]>(() =>
    stateValues[node.props.id] ||
    ((node.props.default_selection ?? []) as number[]).map(keyOf))
  const pickable = selection !== "none"

  const rows: Row[] = useMemo(
    () => rowCells.map((cells, index) => ({ cells, index, key: keyOf(index) })),
    [rowCells, rowKeyList])

  // Where each key sits in this run's rows, which is what puts a selection
  // back in row order and what tells a key whose row is gone from one that is
  // still there.
  const at = useMemo(
    () => new Map(rows.map(row => [row.key, row.index])), [rows])

  // What is drawn as picked is the selection with the current mode applied,
  // mirroring what the Go side hands the page function. The mode can change
  // between runs while this component stays mounted, so a selection made
  // under a wider one must not go on being drawn under a narrower one:
  // dropped outright when the rows are no longer pickable, and cut to one
  // row under single.
  //
  // Which row that is has to be read off the current rows. selected holds
  // the order of the run it was committed in, and the rows can have been
  // reordered or shortened since, which Go settles by resolving the keys
  // against this run's rows before capping. Taking the front of selected
  // instead would paint one row while the page function acted on another.
  const picked = useMemo(() => {
    if (!pickable) {
      return new Set<RowKey>()
    }

    if (selection !== "single") {
      return new Set(selected)
    }

    const first = selected
      .filter(key => at.has(key))
      .sort((a, b) => at.get(a)! - at.get(b)!)[0]

    return new Set(first === undefined ? [] : [first])
  }, [selected, selection, pickable, at])

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

  // commit is the only thing here that talks to the server. The rows are
  // sorted into the order the page function wrote them, whatever order they
  // were picked in, which is what the Go side hands back. A key whose row is
  // gone is dropped here for the same reason the Go side drops it: no row is
  // left for it to name.
  const commit = (keys: RowKey[]) => {
    const next = [...new Set(keys)]
      .filter(key => at.has(key))
      .sort((a, b) => at.get(a)! - at.get(b)!)

    stateValues[node.props.id] = next
    setSelected(next)
    update(keyed
      ? { type: "select", id: node.props.id, keys: next as string[] }
      : { type: "select", id: node.props.id, values: next as number[] })
  }

  const toggleRow = (key: RowKey) => {
    if (selection === "single") {
      // Clicking the picked row again clears it, so a single-select table can
      // be emptied without a modifier key.
      commit(picked.has(key) ? [] : [key])
      return
    }

    commit(picked.has(key)
      ? selected.filter(k => k !== key)
      : [...selected, key])
  }

  // In single mode the row is the only control there is, so it has to be
  // reachable and operable from the keyboard. In multi mode the checkbox
  // already is both, and a focusable row would only add a second tab stop
  // per row without adding anything to do from it.
  const rowProps = (key: RowKey): React.HTMLAttributes<HTMLTableRowElement> =>
    selection !== "single" ? {} : {
      tabIndex: 0,
      onKeyDown: e => {
        if (e.key !== "Enter" && e.key !== " ") {
          return
        }

        // Space would scroll the page, and Enter would submit the form the
        // table may sit in.
        e.preventDefault()
        toggleRow(key)
      },
    }

  // The head checkbox covers every row the search kept, not just the page on
  // screen: paging is how a long table is read, not how it is divided up.
  const allPicked = sorted.length > 0 && sorted.every(row => picked.has(row.key))
  const somePicked = sorted.some(row => picked.has(row.key))

  const toggleAll = () => {
    const inView = sorted.map(row => row.key)
    if (allPicked) {
      const drop = new Set(inView)
      commit(selected.filter(k => !drop.has(k)))
      return
    }

    commit([...selected, ...inView])
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
              <Table.Tr key={String(row.key)}
                {...rowProps(row.key)}
                aria-selected={pickable ? picked.has(row.key) : undefined}
                bg={picked.has(row.key)
                  ? "var(--mantine-color-blue-light)" : undefined}
                style={pickable ? { cursor: "pointer" } : undefined}
                onClick={pickable ? () => toggleRow(row.key) : undefined}>
                {selection === "multi" &&
                  // The click is stopped here so it does not also reach the
                  // row, which would toggle the pick straight back.
                  <Table.Td onClick={e => e.stopPropagation()}>
                    <Checkbox aria-label={`select row ${row.index + 1}`}
                      checked={picked.has(row.key)}
                      onChange={() => toggleRow(row.key)} />
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
