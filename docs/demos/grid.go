package demos

import (
	"github.com/voilelab/toolgui/toolgui/tgcomp"
	"github.com/voilelab/toolgui/toolgui/tgframe"
)

func gridDemo(p *tgframe.Params) error {
	// ANCHOR: demo
	metrics := []struct{ label, value, delta string }{
		{"Revenue", "12.4M", "+12%"},
		{"Users", "8,210", "+3%"},
		{"Churn", "1.8%", "-0.4%"},
		{"Orders", "1,024", "+9%"},
		{"Refunds", "31", "-12%"},
	}

	// As many cells to a row as fit; the fifth wraps, and keeps the width of
	// the four above.
	g := tgcomp.Grid(p.Main, &tgcomp.GridConf{ID: "kpi", MinColWidth: 160})
	for _, m := range metrics {
		tgcomp.Metric(g, m.label, m.value, &tgcomp.MetricConf{Delta: m.delta})
	}
	// ANCHOR_END: demo
	return nil
}
