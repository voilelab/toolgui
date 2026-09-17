package demos

import (
	"github.com/voilelab/toolgui/toolgui/tgcomp"
	"github.com/voilelab/toolgui/toolgui/tgframe"
)

func metricDemo(p *tgframe.Params) error {
	// ANCHOR: demo
	// The second metric is a cost, where growing is the bad news, so its
	// delta is colored the other way round.
	tgcomp.Metric(p.Main, "Revenue", "12.4M",
		&tgcomp.MetricConf{Delta: "+12%"})
	tgcomp.Metric(p.Main, "Cloud spend", "$3.1k",
		&tgcomp.MetricConf{Delta: "+8%", DeltaColorInverse: true})
	// ANCHOR_END: demo
	return nil
}
