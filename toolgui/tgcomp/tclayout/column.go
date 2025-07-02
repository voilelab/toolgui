package tclayout

import (
	"fmt"

	"github.com/voilelab/toolgui/toolgui/tgcomp/tcutil"
	"github.com/voilelab/toolgui/toolgui/tgframe"
)

var _ tgframe.Component = &columnComponent{}
var columnComponentName = "column_component"

type columnComponent struct {
	*tgframe.BaseComponent
	Equal bool `json:"equal"`
}

func newColumnComponent(id string) *columnComponent {
	return &columnComponent{
		BaseComponent: &tgframe.BaseComponent{
			Name: columnComponentName,
			ID:   tcutil.NormalID(columnComponentName, id),
		},
	}
}

// Column create N columns.
func Column(c *tgframe.Container, id string, n uint) []*tgframe.Container {
	if n == 0 {
		panic("number of columns should > 0")
	}

	colsComp := newColumnComponent(id)
	c.AddComponent(colsComp)

	cols := make([]*tgframe.Container, n)
	for i := range n {
		cols[i] = tgframe.NewContainer(fmt.Sprintf("%s_%d", colsComp.ID, i), c.SendNotifyPack)
		c.SendNotifyPack(tgframe.NewNotifyPackCreate(colsComp.ID, cols[i]))
	}

	return cols
}

// Column1 create 1 columns.
func Column1(c *tgframe.Container, id string) *tgframe.Container {
	cols := Column(c, id, 1)
	return cols[0]
}

// Column2 create 2 columns.
func Column2(c *tgframe.Container, id string) (*tgframe.Container, *tgframe.Container) {
	cols := Column(c, id, 2)
	return cols[0], cols[1]
}

// Column3 create 3 columns.
func Column3(c *tgframe.Container, id string) (*tgframe.Container, *tgframe.Container, *tgframe.Container) {
	cols := Column(c, id, 3)
	return cols[0], cols[1], cols[2]
}

// EqColumn create N columns with equal width.
func EqColumn(c *tgframe.Container, id string, n uint) []*tgframe.Container {
	if n == 0 || n > 5 {
		panic("number of columns should be 1, 2, 3, 4, 5")
	}

	colsComp := newColumnComponent(id)
	colsComp.Equal = true
	c.AddComponent(colsComp)

	cols := make([]*tgframe.Container, n)
	for i := range n {
		cols[i] = tgframe.NewContainer(fmt.Sprintf("%s_%d", colsComp.ID, i), c.SendNotifyPack)
		c.SendNotifyPack(tgframe.NewNotifyPackCreate(colsComp.ID, cols[i]))
	}

	return cols
}

// EqColumn1 create 1 column.
func EqColumn1(c *tgframe.Container, id string) *tgframe.Container {
	cols := EqColumn(c, id, 1)
	return cols[0]
}

// EqColumn2 create 2 columns.
func EqColumn2(c *tgframe.Container, id string) (*tgframe.Container, *tgframe.Container) {
	cols := EqColumn(c, id, 2)
	return cols[0], cols[1]
}

// EqColumn3 create 3 columns.
func EqColumn3(c *tgframe.Container, id string) (*tgframe.Container, *tgframe.Container, *tgframe.Container) {
	cols := EqColumn(c, id, 3)
	return cols[0], cols[1], cols[2]
}

// EqColumn4 create 4 columns.
func EqColumn4(c *tgframe.Container, id string) (
	*tgframe.Container, *tgframe.Container, *tgframe.Container, *tgframe.Container) {

	cols := EqColumn(c, id, 4)
	return cols[0], cols[1], cols[2], cols[3]
}

// EqColumn5 create 5 columns.
func EqColumn5(c *tgframe.Container, id string) (
	*tgframe.Container, *tgframe.Container, *tgframe.Container, *tgframe.Container, *tgframe.Container) {

	cols := EqColumn(c, id, 5)
	return cols[0], cols[1], cols[2], cols[3], cols[4]
}
