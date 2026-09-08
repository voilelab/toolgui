package tgcomp

import "github.com/voilelab/toolgui/toolgui/tgcomp/tcdata"

// JSON create a JSON viewer for v.
var JSON = tcdata.JSON

// JSONConf is the configuration for the JSON component.
type JSONConf = tcdata.JSONConf

// Table create a table by heading(head) and values(table).
var Table = tcdata.Table

// TableConf is the configuration for the Table component.
type TableConf = tcdata.TableConf

// DataFrame create a table the user can sort, search and page through, all in
// the browser.
var DataFrame = tcdata.DataFrame

// DataFrameConf is the configuration for the DataFrame component.
type DataFrameConf = tcdata.DataFrameConf

// DataFrameColumnConf is the configuration of one DataFrame column. It is
// named after the component because ColumnConf is the layout Column's.
type DataFrameColumnConf = tcdata.DataFrameColumnConf

// ColumnType is how a DataFrame column's cells are read and sorted.
type ColumnType = tcdata.ColumnType

const (
	// ColumnTypeText sorts the cells as strings.
	ColumnTypeText = tcdata.ColumnTypeText

	// ColumnTypeNumber sorts the cells by their numeric value.
	ColumnTypeNumber = tcdata.ColumnTypeNumber

	// ColumnTypeDatetime sorts the cells by the instant they name.
	ColumnTypeDatetime = tcdata.ColumnTypeDatetime
)

// ColumnAlign is which edge a DataFrame column's cells sit against.
type ColumnAlign = tcdata.ColumnAlign

const (
	// ColumnAlignAuto aligns a number column right, every other one left.
	ColumnAlignAuto = tcdata.ColumnAlignAuto

	// ColumnAlignLeft aligns the cells left.
	ColumnAlignLeft = tcdata.ColumnAlignLeft

	// ColumnAlignCenter centers the cells.
	ColumnAlignCenter = tcdata.ColumnAlignCenter

	// ColumnAlignRight aligns the cells right.
	ColumnAlignRight = tcdata.ColumnAlignRight
)

// ChartKind is the shape a chart is drawn in.
type ChartKind = tcdata.ChartKind

const (
	// ChartKindLine draws one line per series.
	ChartKindLine = tcdata.ChartKindLine

	// ChartKindBar draws one bar per value, grouped by label.
	ChartKindBar = tcdata.ChartKindBar

	// ChartKindArea draws one line per series, filled to the axis.
	ChartKindArea = tcdata.ChartKindArea

	// ChartKindScatter draws one marker per point, on two value axes.
	ChartKindScatter = tcdata.ChartKindScatter
)

// ChartPoint is one point of a scatter chart, on the two value axes.
type ChartPoint = tcdata.ChartPoint

// ChartSeries is one named series of a chart.
type ChartSeries = tcdata.ChartSeries

// ChartConf is the configuration for the chart components.
type ChartConf = tcdata.ChartConf

// Chart create a chart of the kind ChartConf.Kind names.
var Chart = tcdata.Chart

// LineChart create a line chart, one line per series.
var LineChart = tcdata.LineChart

// BarChart create a bar chart, one bar per value grouped by label.
var BarChart = tcdata.BarChart

// AreaChart create an area chart, one filled line per series.
var AreaChart = tcdata.AreaChart

// ScatterChart create a scatter chart, one marker per point.
var ScatterChart = tcdata.ScatterChart
