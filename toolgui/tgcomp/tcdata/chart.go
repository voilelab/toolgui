package tcdata

import (
	"fmt"

	"github.com/voilelab/toolgui/toolgui/tgcomp/tcutil"
	"github.com/voilelab/toolgui/toolgui/tgframe"
)

var _ tgframe.Component = &chartComponent{}
var chartComponentName = "chart_component"

// defaultChartHeight is the CSS height a chart gets when ChartConf leaves it
// empty. A chart fills its container, so it needs an explicit height.
const defaultChartHeight = "300px"

// ChartKind is the shape a chart is drawn in.
type ChartKind int

const (
	// ChartKindLine draws one line per series.
	ChartKindLine ChartKind = iota

	// ChartKindBar draws one bar per value, grouped by label.
	ChartKindBar

	// ChartKindArea draws one line per series, filled to the axis.
	ChartKindArea
)

// String returns the kind as it is named on the wire.
func (k ChartKind) String() string {
	switch k {
	case ChartKindLine:
		return "line"
	case ChartKindBar:
		return "bar"
	case ChartKindArea:
		return "area"
	}

	panic(fmt.Sprintf("unsupported chart kind: %d", int(k)))
}

// ChartSeries is one named series of a chart.
type ChartSeries struct {
	// Name labels the series in the legend and the tooltip.
	Name string `json:"name"`

	// Values holds one value per label, in the same order.
	Values []float64 `json:"values"`

	// Color overrides the theme palette. Any CSS color.
	Color string `json:"color"`
}

// ChartConf is the configuration for the chart components.
type ChartConf struct {
	// Kind is the shape the chart is drawn in, default is ChartKindLine.
	Kind ChartKind

	// Labels are the x axis categories.
	Labels []string

	// Series are the series to draw. Every series needs one value per label.
	Series []ChartSeries

	// Stacked stacks the series on top of each other instead of drawing
	// them side by side.
	Stacked bool

	// Height is the CSS height of the chart (e.g. "300px", "50vh").
	Height string

	// XLabel is the title of the x axis, hidden when empty.
	XLabel string

	// YLabel is the title of the y axis, hidden when empty.
	YLabel string
}

type chartComponent struct {
	*tgframe.BaseComponent
	Kind    string        `json:"kind"`
	Labels  []string      `json:"labels"`
	Series  []ChartSeries `json:"series"`
	Stacked bool          `json:"stacked"`
	Height  string        `json:"height"`
	XLabel  string        `json:"x_label"`
	YLabel  string        `json:"y_label"`
}

func newChartComponent(id string, conf *ChartConf) *chartComponent {
	height := conf.Height
	if height == "" {
		height = defaultChartHeight
	}

	return &chartComponent{
		BaseComponent: &tgframe.BaseComponent{
			Name: chartComponentName,
			ID:   tcutil.NormalID(chartComponentName, id),
		},
		Kind:    conf.Kind.String(),
		Labels:  conf.Labels,
		Series:  conf.Series,
		Stacked: conf.Stacked,
		Height:  height,
		XLabel:  conf.XLabel,
		YLabel:  conf.YLabel,
	}
}

// LineChart create a line chart, one line per series.
// The id has to be unique in the page, and stable across runs so that the
// chart is updated in place instead of redrawn.
func LineChart(c *tgframe.Container, id string, labels []string, series []ChartSeries) {
	ChartWithConf(c, id, &ChartConf{
		Kind:   ChartKindLine,
		Labels: labels,
		Series: series,
	})
}

// BarChart create a bar chart, one bar per value grouped by label.
// The id has to be unique in the page, and stable across runs so that the
// chart is updated in place instead of redrawn.
func BarChart(c *tgframe.Container, id string, labels []string, series []ChartSeries) {
	ChartWithConf(c, id, &ChartConf{
		Kind:   ChartKindBar,
		Labels: labels,
		Series: series,
	})
}

// AreaChart create an area chart, one filled line per series.
// The id has to be unique in the page, and stable across runs so that the
// chart is updated in place instead of redrawn.
func AreaChart(c *tgframe.Container, id string, labels []string, series []ChartSeries) {
	ChartWithConf(c, id, &ChartConf{
		Kind:   ChartKindArea,
		Labels: labels,
		Series: series,
	})
}

// ChartWithConf create a chart with a custom configuration.
func ChartWithConf(c *tgframe.Container, id string, conf *ChartConf) {
	if conf == nil {
		conf = &ChartConf{}
	}

	for _, series := range conf.Series {
		if len(series.Values) != len(conf.Labels) {
			panic(fmt.Sprintf(
				"len of values of series %q should equal to len of labels",
				series.Name))
		}
	}

	comp := newChartComponent(id, conf)
	c.AddComponent(comp)
}
