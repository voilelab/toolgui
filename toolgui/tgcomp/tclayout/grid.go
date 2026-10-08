package tclayout

import (
	"fmt"

	"github.com/voilelab/toolgui/toolgui/tgframe"
)

var _ tgframe.Component = &gridComponent{}

const gridComponentName = "grid_component"

// GridDefaultMinColWidth is the column width, in px, a grid uses when its conf
// gives none.
const GridDefaultMinColWidth = 200

// The gaps a grid may leave between its cells, as the client receives them.
const (
	GridGapNone = "none"
	GridGapXS   = "xs"
	GridGapSM   = "sm"
	GridGapMD   = "md"
	GridGapLG   = "lg"
	GridGapXL   = "xl"
)

type gridComponent struct {
	*tgframe.BaseComponent

	MinColWidth uint   `json:"min_col_width"`
	Gap         string `json:"gap"`
}

func newGridComponent() *gridComponent {
	return &gridComponent{
		BaseComponent: &tgframe.BaseComponent{
			Name: gridComponentName,
		},
	}
}

// gridGap settles what an unset gap means, and rejects one the client cannot
// render.
func gridGap(gap string) string {
	switch gap {
	case "":
		return GridGapMD
	case GridGapNone, GridGapXS, GridGapSM, GridGapMD, GridGapLG, GridGapXL:
		return gap
	}

	panic(fmt.Sprintf("toolgui: unsupported grid gap: %q", gap))
}

// GridConf is the configuration for the Grid component. The container a grid
// hands out derives its id from the grid's; give none and it carries none.
type GridConf struct {
	tgframe.Base

	// MinColWidth is the narrowest a cell may be, in px; it decides how many
	// cells fit a row. 0 means GridDefaultMinColWidth.
	MinColWidth uint

	// Gap is the space between cells: GridGapNone, GridGapXS, GridGapSM,
	// GridGapMD (default), GridGapLG or GridGapXL. Anything else panics.
	Gap string
}

// Grid lays what is written into it out in cells of equal width, as many to a
// row as fit, wrapping onto the next row when the width runs out:
//
//	g := tgcomp.Grid(p.Main)
//	for _, m := range metrics {
//		tgcomp.Metric(g, m.Label, m.Value)
//	}
//
// A short last row keeps the width of the rows above.
func Grid(c *tgframe.Container, conf ...*GridConf) *tgframe.Container {
	cf := tgframe.OneConf("Grid", conf)

	comp := newGridComponent()
	comp.MinColWidth = cf.MinColWidth
	if comp.MinColWidth == 0 {
		comp.MinColWidth = GridDefaultMinColWidth
	}
	comp.Gap = gridGap(cf.Gap)
	tgframe.SetConfIDIn(c, comp, cf)

	gridComp := c.AddComponent(comp)
	return c.AddContainerTo(gridComp, "inner", 0)
}
