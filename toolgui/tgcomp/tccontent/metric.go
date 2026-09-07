package tccontent

import (
	"strings"

	"github.com/voilelab/toolgui/toolgui/tgframe"
)

var _ tgframe.Component = &metricComponent{}
var metricComponentName = "metric_component"

type metricComponent struct {
	*tgframe.BaseComponent
	Label string `json:"label"`
	Value string `json:"value"`
	Delta string `json:"delta"`

	// Direction is "up", "down" or "", and Tone is "positive", "negative" or
	// "". Both are derived from the delta here rather than in the client, so
	// that the rule is stated once and can be tested from Go.
	Direction string `json:"direction"`
	Tone      string `json:"tone"`
}

func newMetricComponent(label, value string) *metricComponent {
	return &metricComponent{
		BaseComponent: &tgframe.BaseComponent{
			Name: metricComponentName,
		},
		Label: label,
		Value: value,
	}
}

// MetricConf is the configuration for the Metric component.
type MetricConf struct {
	tgframe.Base

	// Delta is the change shown under the value, e.g. "+12%" or "-3.2k".
	// Hidden when empty. A delta that starts with "-" is a decrease, anything
	// else an increase.
	Delta string

	// DeltaColorInverse paints an increase red and a decrease green, for a
	// metric where growing is the bad news: cost, latency, churn.
	DeltaColorInverse bool
}

// Metric show a labelled value, with an optional delta under it.
func Metric(c *tgframe.Container, label, value string, conf ...*MetricConf) {
	cf := tgframe.OneConf(conf)

	comp := newMetricComponent(label, value)
	comp.Delta = cf.Delta
	comp.Direction, comp.Tone = deltaDirection(cf.Delta, cf.DeltaColorInverse)
	tgframe.SetConfID(comp, cf)
	c.AddComponent(comp)
}

// deltaDirection reads which way a delta points and whether that is the good
// news. An empty delta points nowhere, and is shown without an arrow or a
// color.
func deltaDirection(delta string, inverse bool) (direction, tone string) {
	if strings.TrimSpace(delta) == "" {
		return "", ""
	}

	down := strings.HasPrefix(strings.TrimSpace(delta), "-")
	if down {
		direction = "down"
	} else {
		direction = "up"
	}

	if down == inverse {
		return direction, "positive"
	}

	return direction, "negative"
}
