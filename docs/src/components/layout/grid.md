# Grid

Grid lays what is written into it out in cells of equal width, as many to a
row as fit, and wraps onto the next row when the width runs out. A row of
cards such as [`Metric`](../content/metric.md) no longer has to be cut into
groups of N and handed to [`EqColumn`](column.md) by hand.

## Usage

```go
func Grid(c *tgframe.Container, conf ...*GridConf) *tgframe.Container
```

* `c`: Parent container.
* `conf`: Optional configuration, at most one.

```go
// GridConf is the configuration for the Grid component.
type GridConf struct {
	tgframe.Base // ID

	// MinColWidth is the narrowest a cell may be, in px; it decides how many
	// cells fit a row. 0 means GridDefaultMinColWidth (200).
	MinColWidth uint

	// Gap is the space between cells: GridGapNone, GridGapXS, GridGapSM,
	// GridGapMD (default), GridGapLG or GridGapXL. Anything else panics.
	Gap string
}
```

The container a grid hands out derives its id from the grid's; give none and
it carries none, and the components inside are still placed by position.

Each component written into the grid is one cell. A short last row keeps the
width of the rows above, and on a screen narrower than `MinColWidth` a cell
takes the full width. A grid works the same inside a [form](../input/form.md),
an [`Empty`](empty.md) slot or a [dialog](dialog.md).

## Example

```go
{{#include ../../../demos/grid.go:demo}}
```

<div data-toolgui-demo="grid"></div>

## Grid or Column?

[`EqColumn`](column.md) gives a fixed number of columns, each a container of
its own — use it to split a page into parts. A grid is for a list of similar
items whose count, or the screen they land on, decides how many fit a row.
