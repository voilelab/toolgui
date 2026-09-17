# Chart

Chart component draws a line, bar or area chart. A scatter chart is drawn from
points rather than labels, and has its own page:
[Scatter Chart](scatter_chart.md).

## API

```go
func Chart(c *tgframe.Container, labels []string, series []ChartSeries, conf ...*ChartConf)
func LineChart(c *tgframe.Container, labels []string, series []ChartSeries, conf ...*ChartConf)
func BarChart(c *tgframe.Container, labels []string, series []ChartSeries, conf ...*ChartConf)
func AreaChart(c *tgframe.Container, labels []string, series []ChartSeries, conf ...*ChartConf)
```

* `c` is the parent container.
* `labels` are the x axis categories.
* `series` are the series to draw. Every series needs one value per label; a
  series of any other length draws an error placeholder instead of the chart
  and fails the run, without stopping the rest of the page.
* `conf` is an optional configuration, at most one.

`LineChart`, `BarChart` and `AreaChart` set `Kind` themselves and ignore what
the conf says; `Chart` follows the conf.

`ChartSeries`:

| Field    | Description                                        |
| -------- | -------------------------------------------------- |
| `Name`   | Shown in the legend and the tooltip.                |
| `Values` | One value per label, in the same order.             |
| `Points` | The `{X, Y}` points of a scatter series, instead of `Values`. |
| `Color`  | Any CSS color, overriding the theme palette.        |

`ChartConf`:

| Field     | Description                                         | Default         |
| --------- | --------------------------------------------------- | --------------- |
| `ID`      | The user specific id, from the embedded `tgframe.Base`. | none         |
| `Kind`    | `ChartKindLine`, `ChartKindBar`, `ChartKindArea` or `ChartKindScatter`. | `ChartKindLine` |
| `Stacked` | Stack the series instead of drawing them side by side. | `false`       |
| `Height`  | CSS height of the chart.                             | `300px`         |
| `XLabel`  | Title of the x axis, hidden when empty.              | none            |
| `YLabel`  | Title of the y axis, hidden when empty.              | none            |

A chart is placed by position like everything else, so it does not need an id
to be updated in place across runs. Give it one when a test or a stylesheet
has to name it, or when the page draws two charts you want to tell apart.

## Examples

### Line

```go
{{#include ../../../demos/chart.go:line}}
```

### Bar

```go
{{#include ../../../demos/chart.go:bar}}
```

### Stacked area

```go
{{#include ../../../demos/chart.go:area}}
```

<div data-toolgui-demo="chart" data-toolgui-demo-height="800">

![chart component](chart.png)

</div>

## Notes

* **Values are `float64`, labels are strings.** A time axis is formatted into
  labels by the page function; there is no time scale.
* **Colors follow the theme.** Series get their color from a palette that is
  stepped for the light and the dark theme, in a fixed order. The palette has
  eight slots and is never cycled, so a chart with more than eight series has
  to set `ChartSeries.Color` on the rest.
* **Every point travels over the websocket on every run of the page
  function.** A few thousand points per chart is the practical ceiling;
  downsample before drawing more than that.
