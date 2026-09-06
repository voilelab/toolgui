# Charts: survey for a first implementation

Survey for [issue #16](https://github.com/voilelab/toolgui/issues/16). It picks
a rendering strategy and sketches the API; it does not implement anything.

## What the repository forces on the design

Five constraints come out of the current code, and they decide most of the
options below.

1. **The core Go module is almost dependency free.** `go.mod` requires only
   `golang.org/x/net`. A pure-Go plotting library (gonum/plot, go-echarts,
   vicanso/go-charts) pulls in a font stack, an image encoder or an HTML
   templating layer, and that lands on every user of `toolgui`.
2. **The wire format is "Go struct → JSON props → React component".** A
   component is a struct embedding `tgframe.BaseComponent`, and
   `toolgui-web/lib/src/components/factory.ts` maps its `name` to a React
   component. Adding a chart is a normal component, not a new mechanism.
3. **Only the browser knows the theme.** `theme` ("light"/"dark") is passed to
   every component as a prop and is flipped client side by
   `ThemeModeButton`. The server never learns it. A canvas does not inherit
   CSS either, so grid, tick and default series colors have to be chosen in
   the frontend.
4. **The desktop build cannot use a CDN.** `app/index.html` loads Bulma and
   Font Awesome from a CDN, but `toolgui-web/wails` depends on `bulma` as an
   npm package because the Wails binary must work offline. A chart library has
   to be an npm dependency of `lib`.
5. **Props are updated in place when the id is stable.** `Forest.createNode`
   reuses an existing node and replaces its props when the id already exists,
   otherwise it creates a new one. A chart that keeps its id updates without
   remounting the canvas; a chart whose id changes every run is destroyed and
   rebuilt each time.

## What Streamlit offers, for reference

Streamlit exposes two layers, and both are load-bearing:

* A narrow high-level layer over a dataframe — `st.line_chart`,
  `st.area_chart`, `st.bar_chart`, `st.scatter_chart`, `st.map`. No chart
  library concepts, just data.
* An escape hatch for everything else — `st.vega_lite_chart`,
  `st.altair_chart`, `st.plotly_chart`, `st.pyplot`, which take a spec or a
  figure from a library the user already knows.

Any first version here should follow the same split: the high-level layer
covers the common case, the escape hatch covers being wrong about what the
common case is.

## Baselines that already work today

Worth stating, because they set the bar a built-in component has to clear.

* **`tgcomp.Iframe` + go-echarts.** go-echarts renders an HTML page; the
  iframe shows it. Full interactivity, zero framework changes, and the chart
  library stays in the *user's* module. But the iframe needs
  `allow-same-origin` to run scripts (see the warning in
  `docs/src/components/misc/iframe.md`), has a fixed CSS height, and has no
  idea about the app theme.
* **`tgcomp.Image` + any pure-Go plotter.** Render a PNG in the page function
  and hand the bytes to `Image`, which already accepts `[]byte` and
  `image.Image`. Static, no tooltip or zoom, but perfect for reports and it
  costs the framework nothing.

Both are recipes, not features: the user has to pick and wire a chart library.
"Charts" as a framework feature means `tgcomp.LineChart(...)` works out of the
box.

## Candidates

| Option | Bundle (min) | Chart types | Interaction | Theme aware | Go deps | Offline/Wails |
| --- | --- | --- | --- | --- | --- | --- |
| **Chart.js v4** | ~60–80 KB tree-shaken (~25–35 KB gz per type) | line, bar, area, scatter, pie, radar | tooltip, legend toggle, hover | manual, per option | none | yes |
| **ECharts** | ~300 KB tree-shaken, ~1 MB full | ~25, incl. maps, heatmap, sankey | rich, zoom/brush built in | built-in themes | none | yes |
| **Vega-Lite** (Streamlit's route) | ~1 MB (vega + vega-lite + embed) | grammar of graphics, anything | selections, tooltip | via config | none | yes |
| **uPlot** | <50 KB | time series, line, area, bar, OHLC | fast, cursor/zoom | manual | none | yes |
| **Server-side Go → PNG** | 0 | depends on library | none | server can't know theme | heavy | yes |

Notes on the ones not recommended:

* **ECharts** is the natural fit if map or sankey charts are wanted soon, but
  it is 4–5x the JavaScript for a line chart, and the win only shows up in
  chart types that are out of scope for a first version.
* **Vega-Lite** gives the most stable public API — a versioned JSON spec means
  the Go side never has to track a rendering library — but the bundle is the
  largest of the lot, and building a spec in Go is much wordier than building
  a Chart.js dataset.
* **uPlot** is the cheapest, and if this were a time-series-only dashboard it
  would win, but no pie/doughnut and a lower-level API make it a poor default.
* **Server-side PNG** is already reachable through `tgcomp.Image` and violates
  constraints 1 and 3. Not worth a component.

## Recommendation

**Chart.js v4 as an npm dependency of `toolgui-web/lib`, with a semantic
(non-Chart.js) wire format and a small high-level Go API.**

It is what the issue links to, it is MIT, canvas based, tree-shakeable, and
the smallest library that still covers pie and scatter. `lib` already ships
`katex`, `react-markdown` and `react-syntax-highlighter`, so it is not the
heaviest thing in the bundle.

### The translation belongs in the frontend, not in Go

The tempting shortcut is to build a Chart.js config object in Go and pass it
through untouched. Two reasons not to, for the high-level API:

* Constraint 3: the theme-dependent parts of the config (grid color, tick
  color, default palette) can only be resolved where the theme is known.
* It would freeze the Chart.js option schema into toolgui's public wire
  format. Swapping renderers later, or upgrading across a Chart.js major,
  would become a breaking change for every user.

So Go sends `{kind, labels, series[], stacked, ...}` and the React component
builds the Chart.js config from it. There is no raw-config escape hatch, now
or later (decided, see "Decisions"): the semantic props are the whole public
surface, which is what keeps the Chart.js schema swappable.

### Go API sketch

Layout components already take an id first (`Box(c, id)`, `Column(c, id, n)`),
and constraint 5 says a chart needs a stable id to update in place rather than
remount. So charts take a **required** id too — unlike `Table`, which uses
`RandID` and is recreated on every run:

```go
package tcdata

type ChartKind int

const (
    ChartKindLine ChartKind = iota
    ChartKindBar
    ChartKindArea
)

// ChartSeries is one named line/bar series.
type ChartSeries struct {
    Name   string
    Values []float64

    // Color overrides the theme palette. Any CSS color.
    Color string
}

type ChartConf struct {
    Kind    ChartKind
    Labels  []string
    Series  []ChartSeries
    Stacked bool
    Height  string // CSS height, default "300px"
    XLabel  string
    YLabel  string
}

func LineChart(c *tgframe.Container, id string, labels []string, series []ChartSeries)
func BarChart(c *tgframe.Container, id string, labels []string, series []ChartSeries)
func AreaChart(c *tgframe.Container, id string, labels []string, series []ChartSeries)
func ChartWithConf(c *tgframe.Container, id string, conf *ChartConf)
```

Usage stays as short as the Streamlit equivalent:

```go
tgcomp.LineChart(p.Main, "sales", []string{"Jan", "Feb", "Mar"},
    []tgcomp.ChartSeries{
        {Name: "2025", Values: []float64{1, 3, 2}},
        {Name: "2026", Values: []float64{2, 2, 5}},
    })
```

`Table` panics on a length mismatch; charts should do the same for
`len(series[i].Values) != len(labels)`.

Values are `float64` and labels are strings, deliberately. A `time.Time`
x-axis would mean a real Chart.js time scale, which needs a date adapter
(`chartjs-adapter-date-fns` and its dependency) in the bundle. Formatting
timestamps into label strings in the page function covers the same charts
without that, so integer and time-typed inputs stay out of the API.

### Wire format

```json
{
  "name": "chart_component",
  "id": "chart_component_sales",
  "kind": "line",
  "labels": ["Jan", "Feb", "Mar"],
  "series": [{"name": "2025", "values": [1, 3, 2], "color": ""}],
  "stacked": false,
  "height": "300px",
  "x_label": "",
  "y_label": ""
}
```

`kind` is a string on the wire even though it is an int in Go, so the JSON
stays readable and a new kind does not renumber anything.

### Frontend component sketch

```tsx
// Register only what is used, so the bundle stays tree-shaken.
Chart.register(LineController, BarController, LineElement, PointElement,
  BarElement, CategoryScale, LinearScale, Filler, Tooltip, Legend)

export function TChart({ node, theme }: Props) {
  const canvasRef = useRef<HTMLCanvasElement>(null)
  const chartRef = useRef<Chart | null>(null)

  useEffect(() => {
    const chart = new Chart(canvasRef.current!, buildConfig(node.props, theme))
    chartRef.current = chart
    return () => chart.destroy()
  }, [])

  useEffect(() => {
    const chart = chartRef.current
    if (!chart) return
    const config = buildConfig(node.props, theme)
    chart.data = config.data
    chart.options = config.options
    chart.update()
  }, [node.props, theme])

  return (
    <div style={{ height: node.props.height, position: 'relative' }}>
      <canvas ref={canvasRef} id={node.props.id} />
    </div>
  )
}
```

Points that are easy to get wrong:

* `destroy()` on unmount, or Chart.js leaks the canvas and its resize
  listener.
* Update in place (`chart.update()`) rather than recreating, so a re-run of the
  page function does not restart the animation.
* `responsive: true` plus `maintainAspectRatio: false` inside a wrapper with an
  explicit height — otherwise the canvas grows on every resize.
* Import from `chart.js` and register explicitly; `chart.js/auto` pulls in
  every controller.

### Theme

`buildConfig` reads `theme` and picks grid/tick colors and the default series
palette. The palette has to be readable on both `theme-light` and
`theme-dark`; the existing CSS in `lib/src/assets/css/shell.css` is the
reference for the surrounding colors. Explicit `ChartSeries.Color` always
wins.

## Scope

**Phase 1 (this issue).** `chart_component` with line, bar and area; the Go
API above; docs and demo. Deliberately no interaction beyond Chart.js's own
tooltip and legend toggle.

**Phase 2.** Scatter, pie and doughnut; per-axis options; horizontal bars;
per-series y-axis.

**Phase 3.** Interaction — a click on a point sending an event back. This
needs a new variant in the `UpdateEvent` union in
`lib/src/app/UpdateEvent.ts` and a matching case on the Go side, so it is a
protocol change, not a component change.

**Out of scope indefinitely.** Maps, 3D, and datasets large enough to need
downsampling or streaming. Every point travels as JSON over the websocket on
every page run, so a few thousand points per chart is the practical ceiling —
that limit belongs in the docs.

## Files a phase 1 change touches

* `toolgui/tgcomp/tcdata/chart.go` — component and constructors.
* `toolgui/tgcomp/data.go` — re-export `LineChart`, `BarChart`, `AreaChart`,
  `ChartWithConf`, `ChartConf`, `ChartSeries`.
* `toolgui-web/lib/src/components/tcdata/chart.tsx` — `TChart`.
* `toolgui-web/lib/src/components/factory.ts` — register `chart_component`.
* `toolgui-web/lib/package.json` — add `chart.js`.
* `cmd/toolgui-demo/main.go` — a row in `DataPage`.
* `docs/src/components/data/chart.md` + `docs/src/SUMMARY.md` +
  `docs/src/components/data/index.md`.
* `toolgui-e2e/cypress/e2e/data.cy.js` — a smoke test. Canvas content is not
  assertable, so check the canvas element and, if the aria fallback is
  enabled, its text.

## Decisions

Settled before implementation:

1. **The id is a required argument.** No hashed default and no remount-per-run
   fallback, so a chart always updates in place.
2. **`float64` values and string labels only.** No time scale and no date
   adapter; a time axis is formatted into labels by the page function.
3. **No raw Chart.js config, ever.** Whatever the high-level API cannot
   express is a gap to close in the API, not to route around. The cost is
   accepted: `ChartConf` will grow options over phases 2 and 3, and users who
   need something genuinely out of scope still have `Iframe` + go-echarts.
