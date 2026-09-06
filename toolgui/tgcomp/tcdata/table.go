package tcdata

import (
	"github.com/voilelab/toolgui/toolgui/tgframe"
)

var _ tgframe.Component = &tableComponent{}
var tableComponentName = "table_component"

type tableComponent struct {
	*tgframe.BaseComponent
	Head  []string   `json:"head"`
	Table [][]string `json:"table"`
}

func newTableComponent(head []string, table [][]string) *tableComponent {
	return &tableComponent{
		BaseComponent: &tgframe.BaseComponent{
			Name: tableComponentName,
		},
		Head:  head,
		Table: table,
	}
}

// Table create a table by heading(head) and values(table).
func Table(c *tgframe.Container, head []string, table [][]string) {
	if len(table) == 0 {
		return
	}

	if len(table[0]) != len(head) {
		panic("len of head should equal to len of table[0]")
	}

	comp := newTableComponent(head, table)
	c.AddComponent(comp)
}
