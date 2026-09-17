package demos

import (
	"github.com/voilelab/toolgui/toolgui/tgcomp"
	"github.com/voilelab/toolgui/toolgui/tgframe"
)

func scatterChartDemo(p *tgframe.Params) error {
	// ANCHOR: demo
	tgcomp.ScatterChart(p.Main,
		[]tgcomp.ChartSeries{
			{Name: "runs", Points: []tgcomp.ChartPoint{
				{X: 1, Y: 3}, {X: 2, Y: 5}, {X: 3, Y: 4},
				{X: 4, Y: 8}, {X: 5, Y: 6},
			}},
		},
		&tgcomp.ChartConf{ID: "demo_scatter"})
	// ANCHOR_END: demo
	return nil
}
