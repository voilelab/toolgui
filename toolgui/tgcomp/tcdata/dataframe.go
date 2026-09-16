package tcdata

import (
	"fmt"
	"slices"
	"strings"

	"github.com/voilelab/toolgui/toolgui/tgcomp/tcutil"
	"github.com/voilelab/toolgui/toolgui/tgframe"
	"github.com/voilelab/toolgui/toolgui/tgutil"
)

var _ tgframe.Component = &dataFrameComponent{}
var dataFrameComponentName = "dataframe_component"

// defaultDataFramePageSize is how many rows a page holds when DataFrameConf
// leaves PageSize empty. Paging, not virtual scrolling, is what keeps a large
// DataFrame smooth: only the current page is ever in the DOM.
const defaultDataFramePageSize = 25

// ColumnType is how a DataFrame column's cells are read, which is what tells
// sorting whether "10" comes before or after "9".
type ColumnType int

const (
	// ColumnTypeText sorts the cells as strings.
	ColumnTypeText ColumnType = iota

	// ColumnTypeNumber sorts the cells by their numeric value, so "10" sorts
	// after "9". Cells that are not numbers sort last.
	ColumnTypeNumber

	// ColumnTypeDatetime sorts the cells by the instant they name. Anything
	// the client can parse as a date works, RFC 3339 most reliably. Cells
	// that are not dates sort last.
	ColumnTypeDatetime
)

// String returns the type as it is named on the wire.
func (t ColumnType) String() string {
	switch t {
	case ColumnTypeText:
		return "text"
	case ColumnTypeNumber:
		return "number"
	case ColumnTypeDatetime:
		return "datetime"
	}

	panic(fmt.Sprintf("unsupported column type: %d", int(t)))
}

// ColumnAlign is which edge a DataFrame column's cells sit against.
type ColumnAlign int

const (
	// ColumnAlignAuto aligns a ColumnTypeNumber column right, and every
	// other column left.
	ColumnAlignAuto ColumnAlign = iota

	// ColumnAlignLeft aligns the cells left.
	ColumnAlignLeft

	// ColumnAlignCenter centers the cells.
	ColumnAlignCenter

	// ColumnAlignRight aligns the cells right.
	ColumnAlignRight
)

// resolve returns the align as it is named on the wire, with ColumnAlignAuto
// settled against the column's type so the client never has to.
func (a ColumnAlign) resolve(t ColumnType) string {
	switch a {
	case ColumnAlignAuto:
		if t == ColumnTypeNumber {
			return "right"
		}
		return "left"
	case ColumnAlignLeft:
		return "left"
	case ColumnAlignCenter:
		return "center"
	case ColumnAlignRight:
		return "right"
	}

	panic(fmt.Sprintf("unsupported column align: %d", int(a)))
}

// SelectionMode is how many rows of a DataFrame the app user may pick.
//
// Picking a row is the one DataFrame interaction that reruns the page
// function: sorting, searching and paging stay in the browser, but a
// selection is an answer the page has to be given.
type SelectionMode int

const (
	// SelectionModeNone leaves the rows unpickable, the default. A DataFrame
	// like this holds no state and always returns an empty selection.
	SelectionModeNone SelectionMode = iota

	// SelectionModeSingle lets one row be picked at a time, by clicking it.
	// Clicking the picked row again clears the selection.
	SelectionModeSingle

	// SelectionModeMulti lets any number of rows be picked, through a
	// checkbox column the table grows on the left.
	SelectionModeMulti
)

// String returns the mode as it is named on the wire.
func (m SelectionMode) String() string {
	switch m {
	case SelectionModeNone:
		return "none"
	case SelectionModeSingle:
		return "single"
	case SelectionModeMulti:
		return "multi"
	}

	panic(fmt.Sprintf("unsupported selection mode: %d", int(m)))
}

// maxSelected is how many rows the mode allows at once, 0 being no limit.
func (m SelectionMode) maxSelected() int {
	if m == SelectionModeSingle {
		return 1
	}
	return 0
}

// DataFrameColumnConf is the configuration of one DataFrame column. It is
// named after the component rather than called ColumnConf because the layout
// Column already has that name.
type DataFrameColumnConf struct {
	// Type is how the cells are read and sorted, default ColumnTypeText.
	Type ColumnType

	// Align is which edge the cells sit against, default ColumnAlignAuto.
	Align ColumnAlign

	// Width is the CSS width of the column (e.g. "8rem", "20%"). The column
	// is sized by its content when empty.
	Width string

	// Hidden drops the column from the table. Its cells are still searched,
	// so a row can be found by a value it does not show.
	Hidden bool
}

// DataFrameConf is the configuration for the DataFrame component. Sorting and
// searching are on unless the conf turns them off, so a DataFrame written
// without a conf is the interactive table the component is for.
type DataFrameConf struct {
	tgframe.Base

	// Sortable lets the user sort by clicking a column head, default on.
	// Set it with SetSortable.
	Sortable *bool

	// Searchable puts a search box above the table that keeps only the rows
	// holding what is typed, default on. Set it with SetSearchable.
	Searchable *bool

	// PageSize is how many rows one page holds, default 25. Set it above the
	// row count to keep every row on one page. Negative panics.
	PageSize int

	// Height is the CSS height the table scrolls past (e.g. "400px"). The
	// page grows the table until it reaches this, and the rows scroll under
	// a pinned head from there. Unbounded when empty.
	Height string

	// ColumnConf configures the columns, one entry per head entry. Empty
	// leaves every column on its defaults; any other length panics.
	ColumnConf []DataFrameColumnConf

	// Selection is how many rows the app user may pick, default
	// SelectionModeNone. Anything else makes DataFrame's return meaningful,
	// and makes picking a row rerun the page function.
	Selection SelectionMode

	// DefaultSelection is what is picked before the app user first touches
	// the table, as indices into rows. It is only read until then. Indices
	// pointing outside rows are dropped, and SelectionModeSingle keeps only
	// the lowest one.
	DefaultSelection []int

	// RowKey is the column whose cells name the rows, so that a selection is
	// remembered by the row it was made on rather than by where that row sat.
	// Left unset, a selection is a position: see [DataFrame]. Set it with
	// SetRowKey.
	//
	// The column has to hold a different value in every row, and to be one of
	// head's; neither failing draws the table.
	RowKey *int
}

// SetRowKey sets RowKey from a column index, which is a pointer so that the
// first column is not what leaving it out means.
func (c *DataFrameConf) SetRowKey(v int) *DataFrameConf {
	c.RowKey = &v
	return c
}

// SetSortable sets Sortable, which is a pointer so that leaving it out means
// on rather than off.
func (c *DataFrameConf) SetSortable(v bool) *DataFrameConf {
	c.Sortable = &v
	return c
}

// SetSearchable sets Searchable, which is a pointer so that leaving it out
// means on rather than off.
func (c *DataFrameConf) SetSearchable(v bool) *DataFrameConf {
	c.Searchable = &v
	return c
}

// dataFrameColumn is one column as the client receives it: every default is
// already settled, so the client only renders.
type dataFrameColumn struct {
	Type   string `json:"type"`
	Align  string `json:"align"`
	Width  string `json:"width"`
	Hidden bool   `json:"hidden"`
}

type dataFrameComponent struct {
	*tgframe.BaseComponent
	Head       []string          `json:"head"`
	Rows       [][]string        `json:"rows"`
	Columns    []dataFrameColumn `json:"columns"`
	Sortable   bool              `json:"sortable"`
	Searchable bool              `json:"searchable"`
	PageSize   int               `json:"page_size"`
	Height     string            `json:"height"`

	Selection        string `json:"selection"`
	DefaultSelection []int  `json:"default_selection"`

	// RowKey is the column the client reads a row's name out of, null when
	// the selection is positional.
	RowKey *int `json:"row_key"`
}

// boolOr reports what an unset conf toggle means.
func boolOr(v *bool, def bool) bool {
	if v == nil {
		return def
	}
	return *v
}

func newDataFrameComponent(head []string, rows [][]string, conf *DataFrameConf) *dataFrameComponent {
	columns := make([]dataFrameColumn, len(head))
	for i := range columns {
		var cc DataFrameColumnConf
		if len(conf.ColumnConf) != 0 {
			cc = conf.ColumnConf[i]
		}

		columns[i] = dataFrameColumn{
			Type:   cc.Type.String(),
			Align:  cc.Align.resolve(cc.Type),
			Width:  cc.Width,
			Hidden: cc.Hidden,
		}
	}

	pageSize := conf.PageSize
	if pageSize == 0 {
		pageSize = defaultDataFramePageSize
	}

	// A DataFrame carries state only once its rows can be picked, so that is
	// the only time it needs an id of its own. The id is derived from the
	// head rather than the rows, so a selection survives the data changing
	// under it; two tables sharing a head are what Conf.ID is for.
	id := ""
	if conf.Selection != SelectionModeNone {
		id = tcutil.HashedID(dataFrameComponentName,
			[]byte(strings.Join(head, "\x00")))
	}

	return &dataFrameComponent{
		BaseComponent: &tgframe.BaseComponent{
			Name: dataFrameComponentName,
			ID:   id,
		},
		Head:       head,
		Rows:       rows,
		Columns:    columns,
		Sortable:   boolOr(conf.Sortable, true),
		Searchable: boolOr(conf.Searchable, true),
		PageSize:   pageSize,
		Height:     conf.Height,
		Selection:  conf.Selection.String(),
		DefaultSelection: normalizeRowSelection(
			conf.DefaultSelection, len(rows), conf.Selection),
		RowKey: conf.RowKey,
	}
}

// storedSelection is the selection as it sits in the state between runs. The
// browser sends both shapes at once: the keys are what a keyed table is read
// back by, and the indices are what an unkeyed one is, so a table that gains
// or loses a RowKey has the other shape already there to fall back on.
type storedSelection struct {
	Indices []int    `json:"indices"`
	Keys    []string `json:"keys"`
}

// readStoredSelection reads what the state holds under id. A plain list of
// indices is read as well as the object the browser sends, because a page
// function is free to write a widget's key itself to seed it, and indices are
// the shape it would reach for -- the same one DataFrameConf.DefaultSelection
// has.
func readStoredSelection(state *tgframe.State, id string) storedSelection {
	var sel storedSelection
	if state.GetObject(id, &sel) == nil {
		return sel
	}

	var idxes []int
	if state.GetObject(id, &idxes) == nil {
		return storedSelection{Indices: idxes}
	}

	return storedSelection{}
}

// resolveRowKeys turns the names of the picked rows into indices into the rows
// there are now. A name no longer in the table is dropped, which is what
// carrying names rather than positions is for: the row it meant is gone, so
// nothing takes its place.
func resolveRowKeys(keys []string, rows [][]string, col int, mode SelectionMode) []int {
	out := []int{}
	if mode == SelectionModeNone {
		return out
	}

	// The keys are unique, which DataFrame refuses to draw a table without,
	// so the first row holding one is the only row holding it.
	at := make(map[string]int, len(rows))
	for i, row := range rows {
		at[row[col]] = i
	}

	for _, key := range keys {
		if idx, ok := at[key]; ok {
			out = append(out, idx)
		}
	}

	slices.Sort(out)
	out = slices.Compact(out)

	if max := mode.maxSelected(); max > 0 && len(out) > max {
		out = out[:max]
	}

	return out
}

// normalizeRowSelection puts a selection into the shape DataFrame promises:
// row order, no duplicates, nothing pointing outside rows, and never nil. The
// mode's cap is applied here too -- the frontend is what keeps the app user
// inside it, but a payload that went past it is trimmed rather than trusted.
func normalizeRowSelection(idxes []int, rowCount int, mode SelectionMode) []int {
	out := []int{}
	if mode == SelectionModeNone {
		return out
	}

	for _, idx := range idxes {
		if idx < 0 || idx >= rowCount {
			continue
		}

		out = append(out, idx)
	}

	slices.Sort(out)
	out = slices.Compact(out)

	if max := mode.maxSelected(); max > 0 && len(out) > max {
		out = out[:max]
	}

	return out
}

// DataFrame create a table the user can sort, search and page through, and
// return the rows the user has picked, as indices into rows. Sorting,
// searching and paging all happen in the browser, so none of them reruns the
// page function.
//
// Every row needs one cell per head entry. [DataFrameColumnConf] carries what
// the cells mean, so the data stays a plain string matrix.
//
// The rows can only be picked once [DataFrameConf.Selection] says so; until
// then the return is always empty. The indices are into rows, the order the
// page function wrote them in, not the order the table happens to show them
// in, and they come back sorted and without duplicates. The result is empty
// rather than nil when nothing is picked.
//
// Left to itself an index is a position and not a row identity, the same
// contract [Select] and [Multiselect] have with their items: when rows changes
// between runs, an index picked against the old data is read against the new
// one, and only an index past the end is dropped.
//
// [DataFrameConf.RowKey] is the way out of that. Point it at the column that
// names the rows and a selection is remembered by the row it was made on: one
// that has moved is still picked, and one that is gone is dropped rather than
// handed to whatever took its place.
//
// [Table] is the static counterpart: reach for it when the rows are few and
// already in the order they should be read in.
func DataFrame(c *tgframe.Container, head []string, rows [][]string,
	conf ...*DataFrameConf) []int {

	cf := tgframe.OneConf("DataFrame", conf)

	if len(head) == 0 {
		c.Fail(tgutil.NewError("a DataFrame needs at least one head entry"))
		return []int{}
	}

	if len(cf.ColumnConf) != 0 && len(cf.ColumnConf) != len(head) {
		c.Fail(tgutil.NewError("len of column conf should equal to len of head"))
		return []int{}
	}

	if cf.PageSize < 0 {
		c.Fail(tgutil.Errorf(
			"page size should not be negative, got %d", cf.PageSize))
		return []int{}
	}

	for i, row := range rows {
		if len(row) != len(head) {
			c.Fail(tgutil.Errorf("len of row %d should equal to len of head", i))
			return []int{}
		}
	}

	if cf.RowKey != nil {
		if *cf.RowKey < 0 || *cf.RowKey >= len(head) {
			c.Fail(tgutil.Errorf(
				"row key should be a column of head, got %d", *cf.RowKey))
			return []int{}
		}

		// A column that names two rows names neither, so picking one of them
		// would pick both. Refused here rather than resolved arbitrarily.
		seen := make(map[string]int, len(rows))
		for i, row := range rows {
			key := row[*cf.RowKey]
			if first, dup := seen[key]; dup {
				c.Fail(tgutil.Errorf(
					"row key column %d should be unique, rows %d and %d are both %q",
					*cf.RowKey, first, i, key))
				return []int{}
			}

			seen[key] = i
		}
	}

	comp := newDataFrameComponent(head, rows, cf)
	tgframe.SetConfID(comp, cf)
	c.AddComponent(comp)

	if cf.Selection == SelectionModeNone {
		return []int{}
	}

	sel := readStoredSelection(c.State, comp.ID)

	// A keyed table is read back by name, so a row that has moved is still
	// the row that was picked and one that has gone takes nothing with it.
	//
	// Names it has none of are not an empty selection but no answer in this
	// shape: every pick made while the table was unkeyed stored none, so a
	// table that has just gained a RowKey reads the indices it did store.
	// The client draws from the same two, in the same order, and the two
	// have to agree.
	if cf.RowKey != nil && len(sel.Keys) != 0 {
		return resolveRowKeys(sel.Keys, rows, *cf.RowKey, cf.Selection)
	}

	if sel.Indices != nil {
		return normalizeRowSelection(sel.Indices, len(rows), cf.Selection)
	}

	// Untouched, so the default stands in. Normalized a second time rather
	// than handing back comp.DefaultSelection: the component is about to be
	// serialized, and the caller owns what it gets back.
	return normalizeRowSelection(cf.DefaultSelection, len(rows), cf.Selection)
}
