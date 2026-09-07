package tccontent

import (
	"encoding/json"
	"testing"

	"github.com/voilelab/toolgui/toolgui/tgframe"
)

// addMetric runs the given call against a container and returns the json the
// container would have sent to the client for the added component.
func addMetric(t *testing.T, add func(c *tgframe.Container)) map[string]any {
	t.Helper()

	var packs []tgframe.NotifyPack
	container := tgframe.NewContainer("test", tgframe.NewState(), func(pack tgframe.NotifyPack) {
		packs = append(packs, pack)
	})

	add(container)

	if len(packs) != 1 {
		t.Fatalf("got %d packs, want 1", len(packs))
	}

	bs, err := json.Marshal(packs[0])
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var out struct {
		Component map[string]any `json:"component"`
	}
	if err := json.Unmarshal(bs, &out); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	return out.Component
}

func TestMetricProps(t *testing.T) {
	props := addMetric(t, func(c *tgframe.Container) {
		Metric(c, "Revenue", "12.4M", &MetricConf{Delta: "+12%"})
	})

	if props["label"] != "Revenue" {
		t.Errorf("label = %v, want Revenue", props["label"])
	}

	if props["value"] != "12.4M" {
		t.Errorf("value = %v, want 12.4M", props["value"])
	}

	if props["delta"] != "+12%" {
		t.Errorf("delta = %v, want +12%%", props["delta"])
	}
}

// The delta's direction is read off its sign, and its tone off the direction
// and whether the metric is one where growing is the bad news.
func TestMetricDeltaDirection(t *testing.T) {
	cases := []struct {
		delta     string
		inverse   bool
		direction string
		tone      string
	}{
		{"+12%", false, "up", "positive"},
		{"12%", false, "up", "positive"},
		{"-3%", false, "down", "negative"},
		{"+8%", true, "up", "negative"},
		{"-8%", true, "down", "positive"},
		{"", false, "", ""},
		{"  ", false, "", ""},
	}

	for _, c := range cases {
		direction, tone := deltaDirection(c.delta, c.inverse)
		if direction != c.direction || tone != c.tone {
			t.Errorf("deltaDirection(%q, %v) = %q, %q, want %q, %q",
				c.delta, c.inverse, direction, tone, c.direction, c.tone)
		}
	}
}
