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

	colsComp := c.AddComponent(newColumnComponent(id))

	cols := make([]*tgframe.Container, n)
	for i := range n {
		cols[i] = c.AddContainerTo(colsComp, fmt.Sprint(i), int(i))
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

// ColumnConf is the configuration for the column components.
type ColumnConf struct {
	// ID is the unique identifier for this column component. The containers it
	// hands out derive their ids from it; give none and they carry none, and
	// the components inside are still placed by position.
	ID string
}

// EqColumn create N columns with equal width. conf may be nil.
func EqColumn(c *tgframe.Container, n uint, conf *ColumnConf) []*tgframe.Container {
	if n == 0 || n > 5 {
		panic("number of columns should be 1, 2, 3, 4, 5")
	}

	comp := &columnComponent{
		BaseComponent: &tgframe.BaseComponent{Name: columnComponentName},
		Equal:         true,
	}
	if conf != nil && conf.ID != "" {
		comp.SetID(conf.ID)
	}

	colsComp := c.AddComponent(comp)

	cols := make([]*tgframe.Container, n)
	for i := range n {
		cols[i] = c.AddContainerTo(colsComp, fmt.Sprint(i), int(i))
	}

	return cols
}

// EqColumn1 create 1 column. conf may be nil.
func EqColumn1(c *tgframe.Container, conf *ColumnConf) *tgframe.Container {
	cols := EqColumn(c, 1, conf)
	return cols[0]
}

// EqColumn2 create 2 columns. conf may be nil.
func EqColumn2(c *tgframe.Container, conf *ColumnConf) (*tgframe.Container, *tgframe.Container) {
	cols := EqColumn(c, 2, conf)
	return cols[0], cols[1]
}

// EqColumn3 create 3 columns. conf may be nil.
func EqColumn3(c *tgframe.Container, conf *ColumnConf) (
	*tgframe.Container, *tgframe.Container, *tgframe.Container) {

	cols := EqColumn(c, 3, conf)
	return cols[0], cols[1], cols[2]
}

// EqColumn4 create 4 columns. conf may be nil.
func EqColumn4(c *tgframe.Container, conf *ColumnConf) (
	*tgframe.Container, *tgframe.Container, *tgframe.Container, *tgframe.Container) {

	cols := EqColumn(c, 4, conf)
	return cols[0], cols[1], cols[2], cols[3]
}

// EqColumn5 create 5 columns. conf may be nil.
func EqColumn5(c *tgframe.Container, conf *ColumnConf) (
	*tgframe.Container, *tgframe.Container, *tgframe.Container, *tgframe.Container,
	*tgframe.Container) {

	cols := EqColumn(c, 5, conf)
	return cols[0], cols[1], cols[2], cols[3], cols[4]
}
