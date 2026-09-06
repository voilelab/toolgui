package tcdata

import (
	"encoding/json"
	"testing"

	"github.com/voilelab/toolgui/toolgui/tgframe"
)

// addChart runs the given call against a container and returns the json the
// container would have sent to the client for the added component.
func addChart(t *testing.T, add func(c *tgframe.Container)) map[string]any {
	t.Helper()

	var packs []tgframe.NotifyPack
	container := tgframe.NewContainer("test", func(pack tgframe.NotifyPack) {
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

func TestLineChartProps(t *testing.T) {
	props := addChart(t, func(c *tgframe.Container) {
		LineChart(c, "sales", []string{"Jan", "Feb"}, []ChartSeries{
			{Name: "2026", Values: []float64{1, 2}},
		})
	})

	if props["id"] != "chart_component_sales" {
		t.Errorf("id = %v, want chart_component_sales", props["id"])
	}

	if props["kind"] != "line" {
		t.Errorf("kind = %v, want line", props["kind"])
	}

	if props["height"] != defaultChartHeight {
		t.Errorf("height = %v, want %v", props["height"], defaultChartHeight)
	}

	if props["stacked"] != false {
		t.Errorf("stacked = %v, want false", props["stacked"])
	}
}

// The id is what keeps a chart updating in place across runs, so the same id
// has to produce the same component id whatever the data is.
func TestChartIDIsStableAcrossData(t *testing.T) {
	first := addChart(t, func(c *tgframe.Container) {
		BarChart(c, "sales", []string{"Jan"}, []ChartSeries{
			{Name: "2026", Values: []float64{1}},
		})
	})

	second := addChart(t, func(c *tgframe.Container) {
		BarChart(c, "sales", []string{"Jan"}, []ChartSeries{
			{Name: "2026", Values: []float64{2}},
		})
	})

	if first["id"] != second["id"] {
		t.Errorf("id = %v and %v, want the same", first["id"], second["id"])
	}
}

func TestChartWithConf(t *testing.T) {
	props := addChart(t, func(c *tgframe.Container) {
		ChartWithConf(c, "traffic", &ChartConf{
			Kind:    ChartKindArea,
			Labels:  []string{"Jan", "Feb"},
			Series:  []ChartSeries{{Name: "hits", Values: []float64{1, 2}}},
			Stacked: true,
			Height:  "500px",
			XLabel:  "month",
			YLabel:  "hits",
		})
	})

	if props["kind"] != "area" {
		t.Errorf("kind = %v, want area", props["kind"])
	}

	if props["height"] != "500px" {
		t.Errorf("height = %v, want 500px", props["height"])
	}

	if props["stacked"] != true {
		t.Errorf("stacked = %v, want true", props["stacked"])
	}

	if props["x_label"] != "month" {
		t.Errorf("x_label = %v, want month", props["x_label"])
	}

	if props["y_label"] != "hits" {
		t.Errorf("y_label = %v, want hits", props["y_label"])
	}
}

func TestChartPanicsOnValueLabelMismatch(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("no panic, want a panic on a series shorter than the labels")
		}
	}()

	addChart(t, func(c *tgframe.Container) {
		LineChart(c, "sales", []string{"Jan", "Feb"}, []ChartSeries{
			{Name: "2026", Values: []float64{1}},
		})
	})
}
