package tcdata

import (
	"slices"
	"strings"
	"testing"

	"github.com/voilelab/toolgui/toolgui/tgcomp/tcutil"
	"github.com/voilelab/toolgui/toolgui/tgframe"
)

// twoByTwo is the smallest data every test here can share.
func twoByTwo() ([]string, [][]string) {
	return []string{"a", "b"}, [][]string{{"1", "2"}, {"3", "4"}}
}

func TestDataFrameDefaults(t *testing.T) {
	head, rows := twoByTwo()
	props := addComponent(t, func(c *tgframe.Container) {
		DataFrame(c, head, rows)
	})

	if props["sortable"] != true {
		t.Errorf("sortable = %v, want true", props["sortable"])
	}

	if props["searchable"] != true {
		t.Errorf("searchable = %v, want true", props["searchable"])
	}

	if props["page_size"] != float64(defaultDataFramePageSize) {
		t.Errorf("page_size = %v, want %v",
			props["page_size"], defaultDataFramePageSize)
	}

	if props["height"] != "" {
		t.Errorf("height = %v, want empty", props["height"])
	}

	columns, ok := props["columns"].([]any)
	if !ok || len(columns) != 2 {
		t.Fatalf("columns = %v, want 2 entries", props["columns"])
	}

	first, _ := columns[0].(map[string]any)
	if first["type"] != "text" || first["align"] != "left" {
		t.Errorf("column 0 = %v, want a left aligned text column", first)
	}
}

func TestDataFrameConf(t *testing.T) {
	head, rows := twoByTwo()
	props := addComponent(t, func(c *tgframe.Container) {
		DataFrame(c, head, rows, (&DataFrameConf{
			ID:       "orders",
			PageSize: 10,
			Height:   "400px",
			ColumnConf: []DataFrameColumnConf{
				{Width: "8rem"},
				{Type: ColumnTypeNumber, Hidden: true},
			},
		}).SetSortable(false).SetSearchable(false))
	})

	if props["id"] != "dataframe_component_orders" {
		t.Errorf("id = %v, want dataframe_component_orders", props["id"])
	}

	if props["sortable"] != false || props["searchable"] != false {
		t.Errorf("sortable = %v, searchable = %v, want both false",
			props["sortable"], props["searchable"])
	}

	if props["page_size"] != float64(10) {
		t.Errorf("page_size = %v, want 10", props["page_size"])
	}

	if props["height"] != "400px" {
		t.Errorf("height = %v, want 400px", props["height"])
	}

	columns := props["columns"].([]any)
	first := columns[0].(map[string]any)
	if first["width"] != "8rem" || first["hidden"] != false {
		t.Errorf("column 0 = %v, want width 8rem and shown", first)
	}

	// A number column is aligned right by ColumnAlignAuto, which the server
	// settles so that the client never sees "auto".
	second := columns[1].(map[string]any)
	if second["type"] != "number" || second["align"] != "right" {
		t.Errorf("column 1 = %v, want a right aligned number column", second)
	}

	if second["hidden"] != true {
		t.Errorf("column 1 hidden = %v, want true", second["hidden"])
	}
}

func TestDataFrameColumnAlign(t *testing.T) {
	for _, tc := range []struct {
		name  string
		align ColumnAlign
		typ   ColumnType
		want  string
	}{
		{"auto text", ColumnAlignAuto, ColumnTypeText, "left"},
		{"auto number", ColumnAlignAuto, ColumnTypeNumber, "right"},
		{"auto datetime", ColumnAlignAuto, ColumnTypeDatetime, "left"},
		{"left over a number", ColumnAlignLeft, ColumnTypeNumber, "left"},
		{"center", ColumnAlignCenter, ColumnTypeText, "center"},
		{"right", ColumnAlignRight, ColumnTypeText, "right"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.align.resolve(tc.typ); got != tc.want {
				t.Errorf("resolve = %q, want %q", got, tc.want)
			}
		})
	}
}

// TestDataFrameWithNoRows pins the defined behaviour of an empty result set:
// the component is still sent, with its head and no rows, so the page shows
// the columns the query would have filled. No rows reaches the client as an
// empty list, not null, so the frontend can iterate it unguarded.
func TestDataFrameWithNoRows(t *testing.T) {
	props := addComponent(t, func(c *tgframe.Container) {
		DataFrame(c, []string{"a", "b"}, nil)
	})

	if rows, ok := props["rows"].([]any); !ok || len(rows) != 0 {
		t.Errorf("rows = %v, want an empty list", props["rows"])
	}

	if head, ok := props["head"].([]any); !ok || len(head) != 2 {
		t.Errorf("head = %v, want 2 entries", props["head"])
	}
}

// TestDataFrameFails covers the checks the run's own data decides: they leave
// an error placeholder rather than taking the page down.
func TestDataFrameFails(t *testing.T) {
	head, rows := twoByTwo()

	for _, tc := range []struct {
		name string
		want string
		add  func(c *tgframe.Container)
	}{
		{"no head", "at least one head entry", func(c *tgframe.Container) {
			DataFrame(c, nil, nil)
		}},
		{"row shorter than head", "len of row 1", func(c *tgframe.Container) {
			DataFrame(c, head, [][]string{{"1", "2"}, {"3"}})
		}},
		{"row longer than head", "len of row 0", func(c *tgframe.Container) {
			DataFrame(c, head, [][]string{{"1", "2", "3"}})
		}},
		{"column conf of the wrong length", "len of column conf",
			func(c *tgframe.Container) {
				DataFrame(c, head, rows, &DataFrameConf{
					ColumnConf: []DataFrameColumnConf{{}},
				})
			}},
		{"negative page size", "should not be negative", func(c *tgframe.Container) {
			DataFrame(c, head, rows, &DataFrameConf{PageSize: -1})
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			msg := failMessage(t, tc.add)
			if !strings.Contains(msg, tc.want) {
				t.Errorf("message = %q, want it to contain %q", msg, tc.want)
			}
		})
	}
}

// TestDataFramePanics keeps the caller mistakes panicking: two confs is a call
// no data can make right.
func TestDataFramePanics(t *testing.T) {
	head, rows := twoByTwo()

	for _, tc := range []struct {
		name string
		add  func(c *tgframe.Container)
	}{
		{"two confs", func(c *tgframe.Container) {
			DataFrame(c, head, rows, &DataFrameConf{}, &DataFrameConf{})
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Errorf("did not panic")
				}
			}()

			container := tgframe.NewContainer("test", tgframe.NewState(),
				func(pack tgframe.NotifyPack) {})
			tc.add(container)
		})
	}
}

// TestDataFrameDoesNotWriteBackToTheCallersConf keeps the conf a caller
// reuses across runs untouched: the page size default is settled on the
// component, not in the conf.
func TestDataFrameDoesNotWriteBackToTheCallersConf(t *testing.T) {
	head, rows := twoByTwo()
	conf := &DataFrameConf{}

	addComponent(t, func(c *tgframe.Container) {
		DataFrame(c, head, rows, conf)
	})

	if conf.PageSize != 0 || conf.Sortable != nil || conf.Searchable != nil {
		t.Errorf("conf = %+v, want it left alone", conf)
	}
}

// selectableHead and selectableRows are the table the selection tests share.
var (
	selectableHead = []string{"host", "region"}
	selectableRows = [][]string{{"web-1", "APAC"}, {"web-2", "EMEA"}, {"db-1", "NA"}}
)

// defaultDataFrameID is the id a selectable DataFrame derives when its conf
// names none. Computed rather than written out so the test says what the id is
// made of -- the head -- instead of pinning a hash.
func defaultDataFrameID(head []string) string {
	return tcutil.HashedID(dataFrameComponentName,
		[]byte(strings.Join(head, "\x00")))
}

// dataFrameSelection runs a DataFrame against state and returns the rows it
// hands back, which is the whole point of a selectable table.
func dataFrameSelection(state *tgframe.State, conf *DataFrameConf) []int {
	container := tgframe.NewContainer("test", state, func(tgframe.NotifyPack) {})
	return DataFrame(container, selectableHead, selectableRows, conf)
}

func TestDataFrameSelectionProps(t *testing.T) {
	for _, tc := range []struct {
		name string
		mode SelectionMode
		want string
	}{
		{"none", SelectionModeNone, "none"},
		{"single", SelectionModeSingle, "single"},
		{"multi", SelectionModeMulti, "multi"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			props := addComponent(t, func(c *tgframe.Container) {
				DataFrame(c, selectableHead, selectableRows,
					&DataFrameConf{Selection: tc.mode})
			})

			if props["selection"] != tc.want {
				t.Errorf("selection = %v, want %v", props["selection"], tc.want)
			}
		})
	}
}

// A DataFrame is unpickable unless its conf says otherwise, so a table written
// before selection existed keeps carrying no state and no id.
func TestDataFrameSelectionDefaultsToNone(t *testing.T) {
	props := addComponent(t, func(c *tgframe.Container) {
		DataFrame(c, selectableHead, selectableRows)
	})

	if props["selection"] != "none" {
		t.Errorf("selection = %v, want none", props["selection"])
	}

	if props["id"] != "" {
		t.Errorf("id = %v, want empty", props["id"])
	}
}

// The id is derived from the head and not the rows, so a selection survives
// the data being refreshed under it.
func TestDataFrameSelectionDerivesIDFromHead(t *testing.T) {
	props := addComponent(t, func(c *tgframe.Container) {
		DataFrame(c, selectableHead, selectableRows,
			&DataFrameConf{Selection: SelectionModeMulti})
	})

	want := defaultDataFrameID(selectableHead)
	if props["id"] != want {
		t.Errorf("id = %v, want %v", props["id"], want)
	}

	// Different rows, same head: the same id, so the pick is not lost.
	other := addComponent(t, func(c *tgframe.Container) {
		DataFrame(c, selectableHead, [][]string{{"db-9", "APAC"}},
			&DataFrameConf{Selection: SelectionModeMulti})
	})

	if other["id"] != want {
		t.Errorf("id after a data change = %v, want %v", other["id"], want)
	}

	// A different head is a different table, so it gets its own id.
	renamed := addComponent(t, func(c *tgframe.Container) {
		DataFrame(c, []string{"host", "zone"}, selectableRows,
			&DataFrameConf{Selection: SelectionModeMulti})
	})

	if renamed["id"] == want {
		t.Errorf("id of another head = %v, want it to differ", renamed["id"])
	}
}

// Two selectable tables sharing a head would collide, which is what Conf.ID
// is for; it wins over the derived one.
func TestDataFrameSelectionConfIDWins(t *testing.T) {
	props := addComponent(t, func(c *tgframe.Container) {
		DataFrame(c, selectableHead, selectableRows, &DataFrameConf{
			ID:        "hosts",
			Selection: SelectionModeMulti,
		})
	})

	if props["id"] != "dataframe_component_hosts" {
		t.Errorf("id = %v, want dataframe_component_hosts", props["id"])
	}
}

// Without a selection mode the return says so whatever the conf asks for, and
// nothing reaches the client either.
func TestDataFrameReturnsNothingWithoutSelection(t *testing.T) {
	got := dataFrameSelection(tgframe.NewState(),
		&DataFrameConf{DefaultSelection: []int{1}})

	if got == nil {
		t.Fatal("DataFrame = nil, want an empty slice")
	}

	if len(got) != 0 {
		t.Fatalf("DataFrame = %v, want empty", got)
	}

	props := addComponent(t, func(c *tgframe.Container) {
		DataFrame(c, selectableHead, selectableRows,
			&DataFrameConf{DefaultSelection: []int{1}})
	})

	if def, ok := props["default_selection"].([]any); !ok || len(def) != 0 {
		t.Errorf("default_selection = %v, want empty", props["default_selection"])
	}
}

func TestDataFrameDefaultSelection(t *testing.T) {
	got := dataFrameSelection(tgframe.NewState(), &DataFrameConf{
		Selection:        SelectionModeMulti,
		DefaultSelection: []int{2, 0, 2},
	})

	if !slices.Equal(got, []int{0, 2}) {
		t.Fatalf("DataFrame = %v, want [0 2]", got)
	}
}

// SelectionModeSingle takes one row, so a default naming more is trimmed
// rather than handed back as a selection the table cannot show.
func TestDataFrameSingleSelectionKeepsOneRow(t *testing.T) {
	got := dataFrameSelection(tgframe.NewState(), &DataFrameConf{
		Selection:        SelectionModeSingle,
		DefaultSelection: []int{2, 1},
	})

	if !slices.Equal(got, []int{1}) {
		t.Fatalf("DataFrame = %v, want [1]", got)
	}
}

// The state holds float64s once the selection has been through JSON, which is
// how it arrives from a real frontend.
func TestDataFrameReadsSelection(t *testing.T) {
	state := tgframe.NewState()
	state.Set(defaultDataFrameID(selectableHead), map[string]any{
		"indices": []float64{2, 0},
		"keys":    []string{},
	})

	got := dataFrameSelection(state, &DataFrameConf{Selection: SelectionModeMulti})
	if !slices.Equal(got, []int{0, 2}) {
		t.Fatalf("DataFrame = %v, want [0 2]", got)
	}
}

// A page function is free to seed a widget's key itself, and a bare list of
// indices is the shape it would reach for -- the one DefaultSelection has. It
// is read as well as the object the browser sends.
func TestDataFrameReadsSelectionWrittenFromGo(t *testing.T) {
	state := tgframe.NewState()
	state.Set(defaultDataFrameID(selectableHead), []int{2, 0})

	got := dataFrameSelection(state, &DataFrameConf{Selection: SelectionModeMulti})
	if !slices.Equal(got, []int{0, 2}) {
		t.Fatalf("DataFrame = %v, want [0 2]", got)
	}
}

// A selection left over from a longer table is dropped rather than clamped,
// the way Select and Multiselect drop one: the row it named is gone.
func TestDataFrameDropsSelectionOutsideRows(t *testing.T) {
	state := tgframe.NewState()
	state.Set(defaultDataFrameID(selectableHead), []int{1, 99, -1})

	got := dataFrameSelection(state, &DataFrameConf{Selection: SelectionModeMulti})
	if !slices.Equal(got, []int{1}) {
		t.Fatalf("DataFrame = %v, want [1]", got)
	}
}

// The default only stands in until the app user has answered, and clearing
// the selection is an answer.
func TestDataFrameEmptySelectionBeatsDefault(t *testing.T) {
	state := tgframe.NewState()
	(&tgframe.EventCustom{
		ID:    defaultDataFrameID(selectableHead),
		Value: map[string]any{"indices": []int{}, "keys": []string{}},
	}).ApplyState(state)

	got := dataFrameSelection(state, &DataFrameConf{
		Selection:        SelectionModeMulti,
		DefaultSelection: []int{1},
	})

	if len(got) != 0 {
		t.Fatalf("DataFrame = %v, want empty", got)
	}
}

// A call that cannot draw its table still hands back a slice, so a caller
// ranging over the result does not have to check the run failed first.
func TestDataFrameSelectionIsNeverNil(t *testing.T) {
	container := tgframe.NewContainer("test", tgframe.NewState(),
		func(tgframe.NotifyPack) {})

	got := DataFrame(container, nil, nil, &DataFrameConf{
		Selection: SelectionModeMulti,
	})

	if got == nil {
		t.Fatal("DataFrame = nil, want an empty slice")
	}
}

// The conf a caller reuses across runs is left alone, the selection included.
func TestDataFrameDoesNotWriteBackTheSelection(t *testing.T) {
	conf := &DataFrameConf{
		Selection:        SelectionModeSingle,
		DefaultSelection: []int{2, 1},
	}

	dataFrameSelection(tgframe.NewState(), conf)

	if !slices.Equal(conf.DefaultSelection, []int{2, 1}) {
		t.Errorf("DefaultSelection = %v, want it left alone", conf.DefaultSelection)
	}
}

// keyedHead names its rows in column 0, so a selection can be remembered by
// the row rather than by where the row sat.
var keyedHead = []string{"host", "region"}

// pickKeyed runs a keyed DataFrame over rows against state.
func pickKeyed(state *tgframe.State, rows [][]string, conf *DataFrameConf) []int {
	c := tgframe.NewContainer("test", state, func(tgframe.NotifyPack) {})
	return DataFrame(c, keyedHead, rows, conf)
}

// keyedConf is a keyed single-select table, which is the shape the identity
// question is sharpest in: one row, and it had better be the right one.
func keyedConf() *DataFrameConf {
	return (&DataFrameConf{Selection: SelectionModeSingle}).SetRowKey(0)
}

// TestDataFrameRowKeyFollowsTheRow is the case that motivated RowKey: pick B
// out of [A B C], drop B's neighbour, and the selection must still be B --
// where a positional selection would have slid onto whatever took the index.
func TestDataFrameRowKeyFollowsTheRow(t *testing.T) {
	abc := [][]string{{"A", "1"}, {"B", "2"}, {"C", "3"}}
	state := tgframe.NewState()

	pickKeyed(state, abc, keyedConf())
	(&tgframe.EventCustom{
		ID:    defaultDataFrameID(keyedHead),
		Value: map[string]any{"indices": []int{1}, "keys": []string{"B"}},
	}).ApplyState(state)

	for _, tc := range []struct {
		name string
		rows [][]string
		want string
	}{
		{"unchanged", abc, "B"},
		{"a row above it removed", [][]string{{"B", "2"}, {"C", "3"}}, "B"},
		{"reordered", [][]string{{"C", "3"}, {"B", "2"}, {"A", "1"}}, "B"},
		{"a row added above it", [][]string{{"Z", "0"}, {"A", "1"}, {"B", "2"}}, "B"},
		{"its cells changed but not its name",
			[][]string{{"A", "1"}, {"B", "9"}, {"C", "3"}}, "B"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := pickKeyed(state, tc.rows, keyedConf())
			if len(got) != 1 {
				t.Fatalf("DataFrame = %v, want one row", got)
			}

			if name := tc.rows[got[0]][0]; name != tc.want {
				t.Errorf("picked %q, want %q", name, tc.want)
			}
		})
	}
}

// The row is gone, so nothing takes its place -- which is the whole point of
// carrying the name rather than the position.
func TestDataFrameRowKeyDropsAVanishedRow(t *testing.T) {
	state := tgframe.NewState()
	(&tgframe.EventCustom{
		ID:    defaultDataFrameID(keyedHead),
		Value: map[string]any{"indices": []int{1}, "keys": []string{"B"}},
	}).ApplyState(state)

	got := pickKeyed(state, [][]string{{"A", "1"}, {"C", "3"}}, keyedConf())
	if len(got) != 0 {
		t.Fatalf("DataFrame = %v, want nothing picked", got)
	}
}

// Without a row key the same data change moves the selection onto another
// row. Pinned so that the difference RowKey makes is a test rather than a
// claim in the docs.
func TestDataFrameWithoutRowKeyMovesWithThePosition(t *testing.T) {
	state := tgframe.NewState()
	state.Set(defaultDataFrameID(keyedHead), []int{1})

	rows := [][]string{{"A", "1"}, {"C", "3"}}
	got := pickKeyed(state, rows, &DataFrameConf{Selection: SelectionModeSingle})
	if len(got) != 1 || rows[got[0]][0] != "C" {
		t.Fatalf("DataFrame = %v, want the row at index 1", got)
	}
}

func TestDataFrameRowKeyMulti(t *testing.T) {
	state := tgframe.NewState()
	(&tgframe.EventCustom{
		ID:    defaultDataFrameID(keyedHead),
		Value: map[string]any{"indices": []int{0, 2}, "keys": []string{"A", "C"}},
	}).ApplyState(state)

	// C moved to the front and A to the back; B is new.
	rows := [][]string{{"C", "3"}, {"B", "2"}, {"A", "1"}}
	got := pickKeyed(state, rows, (&DataFrameConf{
		Selection: SelectionModeMulti,
	}).SetRowKey(0))

	// Row order, as always -- C is index 0 now, A is index 2.
	if !slices.Equal(got, []int{0, 2}) {
		t.Fatalf("DataFrame = %v, want [0 2]", got)
	}

	for _, idx := range got {
		if name := rows[idx][0]; name != "A" && name != "C" {
			t.Errorf("picked %q, want only A and C", name)
		}
	}
}

// A keyed table still caps single at one row, and the row it keeps is the one
// sitting lowest now rather than the one that was lowest when it was picked.
func TestDataFrameRowKeyAppliesTheMode(t *testing.T) {
	state := tgframe.NewState()
	(&tgframe.EventCustom{
		ID:    defaultDataFrameID(keyedHead),
		Value: map[string]any{"indices": []int{0, 2}, "keys": []string{"A", "C"}},
	}).ApplyState(state)

	rows := [][]string{{"C", "3"}, {"A", "1"}}
	got := pickKeyed(state, rows, keyedConf())
	if len(got) != 1 || rows[got[0]][0] != "C" {
		t.Fatalf("DataFrame = %v, want the lowest row, C", got)
	}
}

// The wire carries the column so the client reads names out of the same one.
func TestDataFrameRowKeyProp(t *testing.T) {
	props := addComponent(t, func(c *tgframe.Container) {
		DataFrame(c, keyedHead, [][]string{{"A", "1"}},
			(&DataFrameConf{Selection: SelectionModeSingle}).SetRowKey(1))
	})

	if props["row_key"] != float64(1) {
		t.Errorf("row_key = %v, want 1", props["row_key"])
	}

	unkeyed := addComponent(t, func(c *tgframe.Container) {
		DataFrame(c, keyedHead, [][]string{{"A", "1"}},
			&DataFrameConf{Selection: SelectionModeSingle})
	})

	if unkeyed["row_key"] != nil {
		t.Errorf("row_key = %v, want none", unkeyed["row_key"])
	}
}

// A conf a caller reuses across runs is left alone, the row key included.
func TestDataFrameSetRowKey(t *testing.T) {
	conf := (&DataFrameConf{}).SetRowKey(2)
	if conf.RowKey == nil || *conf.RowKey != 2 {
		t.Fatalf("RowKey = %v, want 2", conf.RowKey)
	}
}

// A table whose key column does not name its rows is refused rather than
// resolved arbitrarily: picking one of two rows sharing a name would pick
// both, and a column out of range names nothing at all.
func TestDataFrameRowKeyFails(t *testing.T) {
	rows := [][]string{{"A", "1"}, {"B", "2"}}

	for _, tc := range []struct {
		name string
		want string
		conf *DataFrameConf
		rows [][]string
	}{
		{"past the head", "should be a column of head",
			(&DataFrameConf{}).SetRowKey(2), rows},
		{"negative", "should be a column of head",
			(&DataFrameConf{}).SetRowKey(-1), rows},
		{"a column that names two rows", "should be unique",
			(&DataFrameConf{}).SetRowKey(1),
			[][]string{{"A", "same"}, {"B", "same"}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			msg := failMessage(t, func(c *tgframe.Container) {
				DataFrame(c, keyedHead, tc.rows, tc.conf)
			})

			if !strings.Contains(msg, tc.want) {
				t.Errorf("message = %q, want it to contain %q", msg, tc.want)
			}
		})
	}
}

// The duplicate message names both rows, so the caller knows where to look.
func TestDataFrameRowKeyDuplicateNamesTheRows(t *testing.T) {
	msg := failMessage(t, func(c *tgframe.Container) {
		DataFrame(c, keyedHead, [][]string{
			{"A", "1"}, {"B", "2"}, {"A", "3"},
		}, (&DataFrameConf{}).SetRowKey(0))
	})

	for _, want := range []string{"rows 0 and 2", `"A"`} {
		if !strings.Contains(msg, want) {
			t.Errorf("message = %q, want it to contain %q", msg, want)
		}
	}
}
