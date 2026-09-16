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
	}
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

	comp := newDataFrameComponent(head, rows, cf)
	tgframe.SetConfID(comp, cf)
	c.AddComponent(comp)

	if cf.Selection == SelectionModeNone {
		return []int{}
	}

	// The state holds whatever the select event landed, a []int written from
	// Go but a []float64 once it has been through JSON; GetObject reads both,
	// and leaves idxes nil for a key holding neither.
	var idxes []int
	if c.State.GetObject(comp.ID, &idxes) != nil {
		idxes = nil
	}

	if idxes == nil {
		// Untouched, so the default stands in. Normalized a second time
		// rather than handing back comp.DefaultSelection: the component is
		// about to be serialized, and the caller owns what it gets back.
		return normalizeRowSelection(cf.DefaultSelection, len(rows), cf.Selection)
	}

	return normalizeRowSelection(idxes, len(rows), cf.Selection)
}
