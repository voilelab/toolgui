package tgcomp

import "github.com/voilelab/toolgui/toolgui/tgcomp/tcdata"

// JSON create a JSON viewer for v.
var JSON = tcdata.JSON

// Table create a table by heading(head) and values(table).
var Table = tcdata.Table

// ChartKind is the shape a chart is drawn in.
type ChartKind = tcdata.ChartKind

const (
	// ChartKindLine draws one line per series.
	ChartKindLine = tcdata.ChartKindLine

	// ChartKindBar draws one bar per value, grouped by label.
	ChartKindBar = tcdata.ChartKindBar

	// ChartKindArea draws one line per series, filled to the axis.
	ChartKindArea = tcdata.ChartKindArea
)

// ChartSeries is one named series of a chart.
type ChartSeries = tcdata.ChartSeries

// ChartConf is the configuration for the chart components.
type ChartConf = tcdata.ChartConf

// LineChart create a line chart, one line per series.
var LineChart = tcdata.LineChart

// BarChart create a bar chart, one bar per value grouped by label.
var BarChart = tcdata.BarChart

// AreaChart create an area chart, one filled line per series.
var AreaChart = tcdata.AreaChart

// ChartWithConf create a chart with a custom configuration.
var ChartWithConf = tcdata.ChartWithConf
