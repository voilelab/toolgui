# DataFrame

DataFrame component displays a table the user can sort, search and page
through, and optionally pick rows from.

Sorting, searching and paging all happen in the browser, so none of them
reruns the page function. `Table` is the static counterpart: reach for it when
the rows are few and already in the order they should be read in, and for
`DataFrame` when the rows are many, or when the reader — not the page
function — should decide the order.

## API

```go
func DataFrame(c *tgframe.Container, head []string, rows [][]string, conf ...*DataFrameConf) []int
```

* `c` is Parent container.
* `head` is the head of table. It cannot be empty.
* `rows` is the body of table. Every row needs one cell per head entry; a row
  of any other length draws an error placeholder instead of the table and
  fails the run, the way a chart does on a series that does not line up with
  its labels.
* `conf` is an optional configuration, at most one.

The return is the rows the user has picked, as indices into `rows`. It is
empty unless [`Selection`](#selection) says the rows can be picked, and empty
rather than nil when nothing is picked.

```go
// DataFrameConf is the configuration for the DataFrame component.
type DataFrameConf struct {
	tgframe.Base // ID

	// Sortable lets the user sort by clicking a column head, default on.
	// Set it with SetSortable.
	Sortable *bool

	// Searchable puts a search box above the table, default on.
	// Set it with SetSearchable.
	Searchable *bool

	// PageSize is how many rows one page holds, default 25.
	PageSize int

	// Height is the CSS height the table scrolls past (e.g. "400px").
	Height string

	// ColumnConf configures the columns, one entry per head entry.
	ColumnConf []DataFrameColumnConf

	// Selection is how many rows the user may pick, default
	// SelectionModeNone.
	Selection SelectionMode

	// RowKeys names the rows, one key per row, so that what is picked is
	// remembered by row and not by position.
	RowKeys []string

	// DefaultSelection is what is picked before the user first touches the
	// table, as indices into rows.
	DefaultSelection []int
}
```

`Sortable` and `Searchable` are pointers so that leaving them out means on,
which is what a `DataFrame` is for. Turn one off through its setter:

```go
conf := (&tgcomp.DataFrameConf{}).SetSearchable(false)
```

Set `PageSize` above the row count to keep every row on one page. A negative
`PageSize` fails the run and draws an error placeholder.

`Height` caps the table rather than fixing it: the page grows the table until
it reaches `Height`, and the rows scroll under a pinned head from there. Left
empty, the table is as tall as its page needs.

### Columns

```go
// DataFrameColumnConf is the configuration of one DataFrame column.
type DataFrameColumnConf struct {
	// Type is how the cells are read and sorted, default ColumnTypeText.
	Type ColumnType

	// Align is which edge the cells sit against, default ColumnAlignAuto.
	Align ColumnAlign

	// Width is the CSS width of the column (e.g. "8rem", "20%").
	Width string

	// Hidden drops the column from the table.
	Hidden bool
}
```

It is named after the component because `ColumnConf` is the layout
[Column](../layout/column.md)'s.

`ColumnConf` is either empty, leaving every column on its defaults, or exactly
as long as `head`. Any other length fails the run and draws an error
placeholder.

The rows stay a plain string matrix; `Type` is what tells the client how to
read them:

| `Type`              | Sorted by                                     |
| ------------------- | --------------------------------------------- |
| `ColumnTypeText`    | the string                                     |
| `ColumnTypeNumber`  | the numeric value, so `"10"` sorts after `"9"` |
| `ColumnTypeDatetime`| the instant named, RFC 3339 most reliably      |

A cell that does not parse as its column's type sorts last, whichever
direction the column is sorted in.

`ColumnAlignAuto` aligns a `ColumnTypeNumber` column right and every other one
left. `ColumnAlignLeft`, `ColumnAlignCenter` and `ColumnAlignRight` say so
outright.

A `Hidden` column is still searched, so a row can be found by a value it does
not show.

### Selection

`Selection` decides how many rows the user may pick, and so whether
`DataFrame`'s return means anything:

| `SelectionMode`        | What the user gets                                       |
| ---------------------- | -------------------------------------------------------- |
| `SelectionModeNone`    | nothing; the rows are not pickable, and the return is empty |
| `SelectionModeSingle`  | one row at a time, picked by clicking it                  |
| `SelectionModeMulti`   | any number of rows, through a checkbox column on the left |

```go
selected := tgcomp.DataFrame(p.Main, head, hosts, &tgcomp.DataFrameConf{
	ID:        "hosts",
	Selection: tgcomp.SelectionModeSingle,
})

if len(selected) != 0 {
	tgcomp.Text(p.Main, "picked "+hosts[selected[0]][0])
}
```

The indices are into `rows`, in the order the page function wrote them — not
the order the table happens to show them in. Sorting and searching only change
what is on screen, so a row picked out of a sorted table still names the row
the page function wrote. They come back sorted and without duplicates, so
`selected` reads the same whatever order the rows were picked in.

In `SelectionModeSingle`, clicking the picked row again clears the selection.
In `SelectionModeMulti`, the checkbox in the head takes every row the search
kept, on whatever page it sits — and gives back only what it took, so a row
picked by hand before the search survives it.

Neither mode needs a mouse. In `SelectionModeSingle` the row is the control,
so each row is a tab stop that <kbd>Enter</kbd> and <kbd>Space</kbd> pick and
clear. In `SelectionModeMulti` the checkbox is already one, so the row is left
out of the tab order rather than made a second stop per row.

`DefaultSelection` is what is picked before the user first touches the table,
and is only read until then: clearing the selection is an answer, and beats
the default from that point on. Indices pointing outside `rows` are dropped,
and `SelectionModeSingle` keeps only the lowest one.

#### Name the rows with `RowKeys`

Without `RowKeys` a selection is a **position**, not a row. If `rows` changes
between runs, an index picked against the old data is read against the new
one. Pick `rows[1]` out of `[A, B, C]`, drop `B`, and the selection is still
`1` — which is now `C`, a row the user never picked. Only an index past the
end of `rows` is dropped. This is the positional contract
[Select](../input/select.md) and [MultiSelect](../input/multiselect.md)
already have with their `items`, and the id being derived from `head` keeps
the selection across a rerun rather than making it safe across a change of
data.

`RowKeys` is how to get out of it, and is what to reach for before acting on a
selection destructively — deleting, submitting, sending. Give one key per row,
whatever the app already calls that row by — a primary key, an id, a path:

```go
keys := make([]string, 0, len(hosts))
for _, host := range hosts {
	keys = append(keys, host[0])
}

selected := tgcomp.DataFrame(p.Main, head, hosts, &tgcomp.DataFrameConf{
	ID:        "hosts",
	Selection: tgcomp.SelectionModeMulti,
	RowKeys:   keys,
})
```

What is picked is then remembered by key. The return is still indices into
`rows` — the same slice the page function just wrote, so `hosts[idx]` reads
the same as ever — but they are resolved from the keys against **this run's**
rows: a row that moved keeps its pick at its new position, and a key whose row
is gone is dropped instead of standing for whatever moved into its place. Drop
`B` out of `[A, B, C]` with `B` picked, and the selection comes back empty.

`RowKeys` is either empty, leaving the selection positional, or exactly as
long as `rows`. Any other length fails the run and draws an error placeholder,
and so does a key used by two rows: two rows answering to one name is the
mistake `RowKeys` exists to rule out. The keys are the app's to give —
`DataFrame` will not guess one out of a column, because no column is
guaranteed unique.

`DefaultSelection` stays positional either way: it is read before the user has
touched anything, when the page function has `rows` in hand.

The other way out is to give the table a fresh `ID` when the data is replaced,
which drops the selection with the old id. That clears the selection rather
than carrying it over, so reach for it when a reload means "start again" and
for `RowKeys` when it does not.

#### Identity

A pickable `DataFrame` holds state, so it needs an id. It derives one from
`head` when the conf names none — from the head rather than the rows, so a
selection survives the data being refreshed underneath it. Two pickable tables
sharing a head on one page would collide, which is what `ID` is for.

An unpickable `DataFrame` holds no state and carries no id at all, so any
number of them can sit on a page.

#### Cost

Picking a row is the one `DataFrame` interaction that reruns the page
function, and the rerun sends `rows` again in full. That is nothing for the
tables selection is usually for, but it is worth knowing before turning it on
for a table of the size [Large tables](#large-tables) measures: there, page
the rows on the server side instead.

## Behaviour

Clicking a column head sorts ascending, clicking it again sorts descending,
and a third click drops the sort and gives the rows back in the order the page
function wrote them.

The search box keeps the rows holding what is typed, matched
case-insensitively against every cell of the row, hidden columns included.

Sorting, searching and paging are all client state. Nothing is sent to the
server, so the page function does not rerun and no other component on the page
is touched. Picking a row is the exception: that is an answer the page has to
be given, so it reruns the page function like any other input component.

With no rows the head is still drawn, over a `No rows` message — which is also
what a search that matches nothing leaves behind.

### Large tables

**Paging, not virtual scrolling, is how a large `DataFrame` stays usable.**
Only `PageSize` rows are ever in the DOM, so what a sort or a keystroke has to
re-render is the page size and not the row count.

Measured on the demo grown to 10,000 rows, each configuration alone on the
page, taking the median of ten samples from click to painted frame:

| | `PageSize` 10 | every row on one page |
| --------------------------- | ------- | --------- |
| rows in the DOM             | 10      | 10,000    |
| page load, first row painted| 660 ms  | 2,922 ms  |
| sort                        | 31 ms   | 1,534 ms  |
| search keystroke            | 31 ms   | 596 ms    |
| page jump                   | 31 ms   | —         |

Scrolling is not what paging rescues: scrolling 4,800 px through all 10,000
rows dropped no frames either way (80 frames, median 16.7 ms, none over
25 ms), because the rows are laid out once and the browser scrolls them on the
compositor. What degrades without paging is **re-rendering** — a sort costs
1.5 s and every search keystroke close to 0.6 s, which is what makes an
unpaged table of this size unusable.

So take the advice above to set `PageSize` above the row count only for tables
of a few hundred rows at most.

Neither number is a server-side bound. The rows are sent whole — 10,000 rows
of this demo's five columns is a 573 KiB pack, the same either way — and are
filtered and sorted in full on every interaction. A table far past 10,000 rows
is worth paging on the server side instead.

## Example

```go
{{#include ../../../demos/dataframe.go:demo}}
```

Picking rows out of one, by key so the pick follows the host:

```go
{{#include ../../../demos/dataframe.go:multi}}
```

Or one row at a time, starting on the first:

```go
{{#include ../../../demos/dataframe.go:single}}
```

<div data-toolgui-demo="dataframe" data-toolgui-demo-height="800">

![dataframe component](dataframe.png)

</div>
