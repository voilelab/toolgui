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

	// DefaultSelection is what is picked before the user first touches the
	// table, as indices into rows.
	DefaultSelection []int

	// RowKey is the column whose cells name the rows, so a selection is
	// remembered by the row rather than by its position. Set it with
	// SetRowKey.
	RowKey *int
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

#### RowKey: naming the rows

Left to itself a selection is a **position**. If `rows` changes between runs,
an index picked against the old data is read against the new one: pick
`rows[1]` out of `[A, B, C]`, drop `B`, and the selection is still `1` — which
is now `C`, a row the user never picked. Only an index past the end of `rows`
is dropped. That is the positional contract
[Select](../input/select.md) and [Multiselect](../input/multiselect.md) have
with their `items`, and it is fine for a table whose rows keep their order.

`RowKey` is the way out. Point it at the column that names the rows and a
selection is remembered by the row it was made on:

```go
selected := tgcomp.DataFrame(p.Main, head, queues,
	(&tgcomp.DataFrameConf{
		ID:        "queues",
		Selection: tgcomp.SelectionModeSingle,
	}).SetRowKey(0))
```

| What happens to the picked row | Positional | With `RowKey` |
| --- | --- | --- |
| a row above it is removed | slides onto its neighbour | stays on the row |
| the rows are reordered | stays on the position | stays on the row |
| its other cells change | stays on the row | stays on the row |
| the row itself is removed | slides onto whatever took the index | dropped |

It is a pointer so that leaving it out does not mean the first column; set it
through `SetRowKey`. The column has to be one of `head`'s and has to hold a
different value in every row — a column naming two rows names neither, and
picking one of them would pick both, so the run fails and draws an error
placeholder rather than resolving it arbitrarily.

The return is still indices into the current `rows` either way. `RowKey`
changes what is *remembered*, not what is handed back, so `rows[selected[0]]`
stays the way to read the picked row.

`DefaultSelection` stays positional even with a `RowKey` set: the page
function writes it fresh against the rows it just supplied, so there is no
older data for it to be read against.

So: reach for `RowKey` whenever the rows are re-queried, filtered by another
widget, or refreshed, and especially before acting on a selection
destructively — deleting, submitting, sending.

#### Identity

A pickable `DataFrame` holds state, so it needs an id. It derives one from
`head` when the conf names none — from the head rather than the rows, so the
selection is not thrown away every time the data is refreshed. Two pickable
tables sharing a head on one page would collide, which is what `ID` is for.

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
tgcomp.DataFrame(p.Main,
	[]string{"Order", "Ordered", "Region", "Item", "Amount"},
	orders,
	&tgcomp.DataFrameConf{
		ID:       "orders",
		PageSize: 10,
		ColumnConf: []tgcomp.DataFrameColumnConf{
			{Width: "9rem"},
			{Type: tgcomp.ColumnTypeDatetime},
			{},
			{},
			{Type: tgcomp.ColumnTypeNumber},
		},
	})
```

Picking rows out of one:

```go
selected := tgcomp.DataFrame(p.Main,
	[]string{"Host", "Region", "Status"},
	hosts,
	&tgcomp.DataFrameConf{
		ID:        "hosts",
		Selection: tgcomp.SelectionModeMulti,
	})

names := []string{}
for _, idx := range selected {
	names = append(names, hosts[idx][0])
}

tgcomp.Text(p.Main, "Selected: "+strings.Join(names, ", "))
```

![dataframe component](dataframe.png)
