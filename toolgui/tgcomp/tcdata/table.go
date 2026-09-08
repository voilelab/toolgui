package tcdata

import (
	"github.com/voilelab/toolgui/toolgui/tgframe"
	"github.com/voilelab/toolgui/toolgui/tgutil"
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

// TableConf is the configuration for the Table component.
type TableConf struct {
	tgframe.Base
}

// Table create a table by heading(head) and values(table).
func Table(c *tgframe.Container, head []string, table [][]string, conf ...*TableConf) {
	cf := tgframe.OneConf("Table", conf)

	if len(table) == 0 {
		return
	}

	if len(table[0]) != len(head) {
		c.Fail(tgutil.NewError("len of head should equal to len of table[0]"))
		return
	}

	comp := newTableComponent(head, table)
	tgframe.SetConfID(comp, cf)
	c.AddComponent(comp)
}
