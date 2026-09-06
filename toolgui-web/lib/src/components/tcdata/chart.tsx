import React, { useEffect, useRef } from "react"

import {
  BarController,
  BarElement,
  CategoryScale,
  Chart,
  Filler,
  Legend,
  LineController,
  LineElement,
  LinearScale,
  PointElement,
  Tooltip,
} from "chart.js"

import { Props } from "../component_interface"

// Register only what the supported kinds need. chart.js/auto would pull in
// every controller and scale instead.
Chart.register(
  BarController, BarElement, LineController, LineElement, PointElement,
  CategoryScale, LinearScale, Filler, Legend, Tooltip)

// Categorical palette, validated for both themes. Slots are handed out in
// order and never cycled: a chart with more series than slots has to name its
// own colors.
const palettes: { [theme: string]: string[] } = {
  light: ['#2a78d6', '#eb6834', '#1baf7a', '#eda100',
    '#e87ba4', '#008300', '#4a3aa7', '#e34948'],
  dark: ['#3987e5', '#d95926', '#199e70', '#c98500',
    '#d55181', '#008300', '#9085e9', '#e66767'],
}

// Grid, axis and label colors. Recessive by design: one step off the surface.
const chromes: { [theme: string]: { grid: string, tick: string, surface: string } } = {
  light: { grid: '#e1e0d9', tick: '#898781', surface: '#ffffff' },
  dark: { grid: '#2c2c2a', tick: '#898781', surface: '#14161a' },
}

// Points stop being readable once they crowd; past this many labels the line
// is drawn bare and the markers only show on hover.
const maxPointLabels = 30

// resolveTheme reports the theme the page is actually painted in. The theme
// prop is only set once the viewer has used the toggle, while Bulma also
// follows prefers-color-scheme, so the painted background is the better
// source and it doubles as the surface color.
function resolveTheme(theme: string): { theme: string, surface: string } {
  const painted = paintedBackground()
  if (painted) {
    return { theme: isDark(painted) ? 'dark' : 'light', surface: painted }
  }

  const name = theme === 'dark' ? 'dark' : 'light'
  return { theme: name, surface: chromes[name].surface }
}

// paintedBackground returns the page background as an rgb() string, or ''
// when nothing painted one (a test renderer, an unstyled page). Bulma paints
// the html element, so that is what has to be read; body stays transparent.
function paintedBackground(): string {
  if (typeof window === 'undefined') {
    return ''
  }

  for (const element of [document.documentElement, document.body]) {
    if (!element) {
      continue
    }

    const color = window.getComputedStyle(element).backgroundColor
    if (color && parseRGB(color)) {
      return color
    }
  }

  return ''
}

function parseRGB(color: string): number[] | null {
  const match = /^rgba?\(([^)]+)\)$/.exec(color.trim())
  if (!match) {
    return null
  }

  const parts = match[1].split(/[,/\s]+/).map(parseFloat)
  if (parts.length < 3 || parts.some(isNaN)) {
    return null
  }

  // A fully transparent body shows whatever is behind it, so it says nothing
  // about the surface.
  if (parts.length > 3 && parts[3] === 0) {
    return null
  }

  return parts
}

function isDark(color: string): boolean {
  const rgb = parseRGB(color)
  if (!rgb) {
    return false
  }

  // Rec. 601 luma, enough to tell a dark surface from a light one.
  return (0.299 * rgb[0] + 0.587 * rgb[1] + 0.114 * rgb[2]) < 128
}

// wash returns the series color at the opacity an area fill wants: a tint of
// the hue, never a saturated block.
function wash(color: string): string {
  const hex = /^#([0-9a-f]{3}|[0-9a-f]{6})$/i.exec(color)
  if (!hex) {
    return `color-mix(in srgb, ${color} 12%, transparent)`
  }

  const digits = hex[1].length === 3
    ? hex[1].split('').map(d => d + d).join('')
    : hex[1]

  const r = parseInt(digits.slice(0, 2), 16)
  const g = parseInt(digits.slice(2, 4), 16)
  const b = parseInt(digits.slice(4, 6), 16)
  return `rgba(${r}, ${g}, ${b}, 0.12)`
}

function seriesColor(series: any, index: number, theme: string): string {
  if (series.color) {
    return series.color
  }

  const palette = palettes[theme]
  if (index >= palette.length) {
    throw new Error(
      `chart has more than ${palette.length} series, ` +
      'set ChartSeries.Color on the rest')
  }

  return palette[index]
}

function buildConfig(props: any, theme: string, surface: string): any {
  const area = props.kind === 'area'
  const bar = props.kind === 'bar'
  const chrome = chromes[theme]
  const showPoints = props.labels.length <= maxPointLabels

  const datasets = props.series.map((series: any, index: number) => {
    const color = seriesColor(series, index, theme)

    if (bar) {
      return {
        label: series.name,
        data: series.values,
        backgroundColor: color,
        // Cap the thickness and leave the band's air; round the growing end
        // and keep the baseline square.
        maxBarThickness: 24,
        borderRadius: 4,
        borderSkipped: 'start',
      }
    }

    return {
      label: series.name,
      data: series.values,
      borderColor: color,
      backgroundColor: area ? wash(color) : color,
      borderWidth: 2,
      borderJoinStyle: 'round',
      borderCapStyle: 'round',
      // Stacked areas fill down to the series below, except the bottom one,
      // which has nothing under it. Everything else fills to the axis.
      fill: area ? (props.stacked && index > 0 ? '-1' : 'origin') : false,
      pointRadius: showPoints ? 4 : 0,
      pointHoverRadius: 5,
      // A ring in the surface color keeps a marker legible where it crosses
      // another line.
      pointBorderColor: surface,
      pointBorderWidth: 2,
      pointBackgroundColor: color,
    }
  })

  return {
    type: bar ? 'bar' : 'line',
    data: { labels: props.labels, datasets: datasets },
    options: {
      responsive: true,
      // The wrapper owns the height, so the canvas must not keep a ratio.
      maintainAspectRatio: false,
      interaction: bar
        ? { mode: 'nearest', intersect: true }
        : { mode: 'index', intersect: false },
      plugins: {
        // One series is named by the page around it; two or more need the
        // legend, so identity is never carried by color alone.
        legend: {
          display: props.series.length > 1,
          position: 'top',
          align: 'start',
          labels: { usePointStyle: true, boxWidth: 8, color: chrome.tick },
        },
      },
      scales: {
        x: {
          stacked: props.stacked,
          grid: { display: false },
          border: { color: chrome.grid },
          ticks: { color: chrome.tick },
          title: {
            display: !!props.x_label,
            text: props.x_label,
            color: chrome.tick,
          },
        },
        y: {
          stacked: props.stacked,
          // Bars and area fills encode magnitude by their length, so a
          // truncated baseline would overstate the differences. It also keeps
          // the fill target inside the plot, which is what makes it visible.
          beginAtZero: bar || area,
          grid: { color: chrome.grid },
          border: { display: false },
          ticks: { color: chrome.tick },
          title: {
            display: !!props.y_label,
            text: props.y_label,
            color: chrome.tick,
          },
        },
      },
    },
  }
}

export function TChart({ node, theme }: Props) {
  const canvasRef = useRef<HTMLCanvasElement>(null)
  const chartRef = useRef<Chart | null>(null)

  // node.props is replaced on every run of the page function, so the chart is
  // updated in place instead of being torn down and rebuilt.
  useEffect(() => {
    const canvas = canvasRef.current
    if (!canvas) {
      return
    }

    const resolved = resolveTheme(theme)
    const config = buildConfig(node.props, resolved.theme, resolved.surface)

    const chart = chartRef.current

    // chart.js cannot switch an existing chart between line and bar, so a
    // kind change on the same id is the one case that has to be rebuilt.
    if (chart && (chart.config as any).type !== config.type) {
      chart.destroy()
      chartRef.current = null
    }

    if (!chartRef.current) {
      chartRef.current = new Chart(canvas, config)
      return
    }

    chartRef.current.data = config.data
    chartRef.current.options = config.options
    chartRef.current.update()
  }, [node.props, theme])

  // Only on unmount: chart.js keeps a resize listener alive otherwise.
  useEffect(() => {
    return () => {
      chartRef.current?.destroy()
      chartRef.current = null
    }
  }, [])

  return (
    <div className="block" style={{ height: node.props.height, position: 'relative' }}>
      <canvas ref={canvasRef} id={node.props.id} role="img"
        aria-label={chartLabel(node.props)} />
    </div>
  )
}

function chartLabel(props: any): string {
  const names = props.series.map((series: any) => series.name).join(', ')
  return `${props.kind} chart of ${names}`
}
