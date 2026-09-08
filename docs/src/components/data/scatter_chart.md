# Scatter Chart

Scatter Chart component draws one marker per point, on two value axes.

It is the one chart that takes points rather than labels and values: both of
its axes are value axes, so a point carries its own x. The rest of
[Chart](chart.md) applies unchanged — the conf, the palette, the notes.

## API

```go
func ScatterChart(c *tgframe.Container, series []ChartSeries, conf ...*ChartConf)
```

* `c` is the parent container.
* `series` are the series to draw. Every series needs its `Points`, and no
  `Values`; anything else draws an error placeholder instead of the chart and
  fails the run, without stopping the rest of the page.
* `conf` is an optional configuration, at most one. `ScatterChart` sets `Kind`
  itself and ignores what the conf says.

`ChartPoint`:

| Field | Description         |
| ----- | ------------------- |
| `X`   | Position on x axis. |
| `Y`   | Position on y axis. |

## Example

```go
tgcomp.ScatterChart(p.Main, []tgcomp.ChartSeries{
	{Name: "runs", Points: []tgcomp.ChartPoint{
		{X: 1, Y: 3}, {X: 2, Y: 5}, {X: 3, Y: 4},
		{X: 4, Y: 8}, {X: 5, Y: 6},
	}},
})
```

![scatter chart component](scatter_chart.png)

Axis titles and a height come from the conf, the same as the other kinds:

```go
tgcomp.ScatterChart(p.Main,
	[]tgcomp.ChartSeries{{Name: "p95", Points: points}},
	&tgcomp.ChartConf{
		XLabel: "concurrency",
		YLabel: "ms",
	})
```
