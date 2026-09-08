package tcdata

import (
	"fmt"

	"github.com/voilelab/toolgui/toolgui/tgframe"
	"github.com/voilelab/toolgui/toolgui/tgutil"
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

	// ChartKindScatter draws one marker per point, on two value axes.
	ChartKindScatter
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
	case ChartKindScatter:
		return "scatter"
	}

	panic(fmt.Sprintf("unsupported chart kind: %d", int(k)))
}

// ChartPoint is one point of a scatter chart, on the two value axes.
type ChartPoint struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// ChartSeries is one named series of a chart.
type ChartSeries struct {
	// Name labels the series in the legend and the tooltip.
	Name string `json:"name"`

	// Values holds one value per label, in the same order. Every kind but
	// ChartKindScatter is drawn from it.
	Values []float64 `json:"values"`

	// Points holds the {x, y} points of a ChartKindScatter series, which has
	// no labels to line its values up with. Omitted from the wire for the
	// other kinds, so their packs are unchanged.
	Points []ChartPoint `json:"points,omitempty"`

	// Color overrides the theme palette. Any CSS color.
	Color string `json:"color"`
}

// ChartConf is the configuration for the chart components. The data is not in
// here: labels and series are what a chart is for, so every entry point takes
// them positionally and this carries only the presentation.
type ChartConf struct {
	tgframe.Base

	// Kind is the shape the chart is drawn in, default is ChartKindLine.
	// [LineChart], [BarChart], [AreaChart] and [ScatterChart] set it
	// themselves.
	Kind ChartKind

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

func newChartComponent(labels []string, series []ChartSeries, conf *ChartConf) *chartComponent {
	height := conf.Height
	if height == "" {
		height = defaultChartHeight
	}

	return &chartComponent{
		BaseComponent: &tgframe.BaseComponent{
			Name: chartComponentName,
		},
		Kind:    conf.Kind.String(),
		Labels:  labels,
		Series:  series,
		Stacked: conf.Stacked,
		Height:  height,
		XLabel:  conf.XLabel,
		YLabel:  conf.YLabel,
	}
}

// Chart create a chart of the kind [ChartConf.Kind] names, default a line
// chart. Every series needs one value per label.
func Chart(c *tgframe.Container, labels []string, series []ChartSeries, conf ...*ChartConf) {
	chart(c, labels, series, tgframe.OneConf("Chart", conf), nil)
}

// LineChart create a line chart, one line per series.
func LineChart(c *tgframe.Container, labels []string, series []ChartSeries, conf ...*ChartConf) {
	kind := ChartKindLine
	chart(c, labels, series, tgframe.OneConf("LineChart", conf), &kind)
}

// BarChart create a bar chart, one bar per value grouped by label.
func BarChart(c *tgframe.Container, labels []string, series []ChartSeries, conf ...*ChartConf) {
	kind := ChartKindBar
	chart(c, labels, series, tgframe.OneConf("BarChart", conf), &kind)
}

// AreaChart create an area chart, one filled line per series.
func AreaChart(c *tgframe.Container, labels []string, series []ChartSeries, conf ...*ChartConf) {
	kind := ChartKindArea
	chart(c, labels, series, tgframe.OneConf("AreaChart", conf), &kind)
}

// ScatterChart create a scatter chart, one marker per point. It takes no
// labels: both of a scatter chart's axes are value axes, so a point carries
// its own x, in [ChartSeries.Points] rather than Values.
func ScatterChart(c *tgframe.Container, series []ChartSeries, conf ...*ChartConf) {
	kind := ChartKindScatter
	chart(c, nil, series, tgframe.OneConf("ScatterChart", conf), &kind)
}

// chart adds the chart component. kind is what the entry point is named after,
// and overrides whatever the conf says; [Chart] passes nil and leaves the
// conf's own kind alone. The conf is copied rather than written through: the
// caller owns it and may reuse it across runs.
func chart(c *tgframe.Container, labels []string, series []ChartSeries,
	conf *ChartConf, kind *ChartKind) {

	cf := *conf
	if kind != nil {
		cf.Kind = *kind
	}

	for _, s := range series {
		if cf.Kind == ChartKindScatter {
			if len(s.Values) != 0 {
				c.Fail(tgutil.Errorf(
					"series %q of a scatter chart is drawn from points, not values",
					s.Name))
				return
			}
			continue
		}

		if len(s.Points) != 0 {
			c.Fail(tgutil.Errorf(
				"series %q holds points, which only a scatter chart draws",
				s.Name))
			return
		}

		if len(s.Values) != len(labels) {
			c.Fail(tgutil.Errorf(
				"len of values of series %q should equal to len of labels",
				s.Name))
			return
		}
	}

	comp := newChartComponent(labels, series, &cf)
	tgframe.SetConfID(comp, conf)
	c.AddComponent(comp)
}
