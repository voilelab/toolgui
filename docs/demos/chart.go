package demos

import (
	"github.com/voilelab/toolgui/toolgui/tgcomp"
	"github.com/voilelab/toolgui/toolgui/tgframe"
)

func lineChartDemo(p *tgframe.Params) error {
	// ANCHOR: line
	tgcomp.LineChart(p.Main,
		[]string{"Mon", "Tue", "Wed", "Thu", "Fri"},
		[]tgcomp.ChartSeries{
			{Name: "visits", Values: []float64{12, 19, 9, 24, 17}},
			{Name: "signups", Values: []float64{3, 7, 4, 9, 6}},
		},
		&tgcomp.ChartConf{ID: "demo_line"})
	// ANCHOR_END: line
	return nil
}

func barChartDemo(p *tgframe.Params) error {
	// ANCHOR: bar
	tgcomp.BarChart(p.Main,
		[]string{"Go", "Rust", "Python"},
		[]tgcomp.ChartSeries{
			{Name: "stars", Values: []float64{31, 24, 47}},
		},
		&tgcomp.ChartConf{ID: "demo_bar"})
	// ANCHOR_END: bar
	return nil
}

func areaChartDemo(p *tgframe.Params) error {
	// ANCHOR: area
	tgcomp.AreaChart(p.Main,
		[]string{"Q1", "Q2", "Q3", "Q4"},
		[]tgcomp.ChartSeries{
			{Name: "cloud", Values: []float64{4, 6, 5, 9}},
			{Name: "desktop", Values: []float64{2, 3, 4, 4}},
		},
		&tgcomp.ChartConf{
			ID:      "demo_area",
			Stacked: true,
			YLabel:  "revenue",
		})
	// ANCHOR_END: area
	return nil
}
