package tcdata

import (
	"testing"

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
			Base:     tgframe.Base{ID: "orders"},
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
// the columns the query would have filled.
func TestDataFrameWithNoRows(t *testing.T) {
	props := addComponent(t, func(c *tgframe.Container) {
		DataFrame(c, []string{"a", "b"}, nil)
	})

	if props["rows"] != nil {
		t.Errorf("rows = %v, want none", props["rows"])
	}

	if head, ok := props["head"].([]any); !ok || len(head) != 2 {
		t.Errorf("head = %v, want 2 entries", props["head"])
	}
}

func TestDataFramePanics(t *testing.T) {
	head, rows := twoByTwo()

	for _, tc := range []struct {
		name string
		add  func(c *tgframe.Container)
	}{
		{"no head", func(c *tgframe.Container) {
			DataFrame(c, nil, nil)
		}},
		{"row shorter than head", func(c *tgframe.Container) {
			DataFrame(c, head, [][]string{{"1", "2"}, {"3"}})
		}},
		{"row longer than head", func(c *tgframe.Container) {
			DataFrame(c, head, [][]string{{"1", "2", "3"}})
		}},
		{"column conf of the wrong length", func(c *tgframe.Container) {
			DataFrame(c, head, rows, &DataFrameConf{
				ColumnConf: []DataFrameColumnConf{{}},
			})
		}},
		{"negative page size", func(c *tgframe.Container) {
			DataFrame(c, head, rows, &DataFrameConf{PageSize: -1})
		}},
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

func TestColumnTypeString(t *testing.T) {
	for _, tc := range []struct {
		typ  ColumnType
		want string
	}{
		{ColumnTypeText, "text"},
		{ColumnTypeNumber, "number"},
		{ColumnTypeDatetime, "datetime"},
	} {
		if got := tc.typ.String(); got != tc.want {
			t.Errorf("String = %q, want %q", got, tc.want)
		}
	}

	defer func() {
		if recover() == nil {
			t.Errorf("an unsupported type did not panic")
		}
	}()
	_ = ColumnType(9).String()
}
