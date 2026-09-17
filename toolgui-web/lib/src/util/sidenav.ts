import { getStoredValue, setStoredValue } from './storage'

const STORAGE_KEY = 'sidenav_collapsed'

// The column starts expanded, the way it was before the toggle existed, for
// anyone who has never touched the toggle.
const DEFAULT_COLLAPSED = false

// storedNavCollapsed is the choice the visitor made last time, or undefined
// if they never made one — a distinction the default above depends on.
function storedNavCollapsed(): boolean | undefined {
  const stored = getStoredValue(STORAGE_KEY)
  if (stored === 'true') {
    return true
  }

  if (stored === 'false') {
    return false
  }

  return undefined
}

// initialNavCollapsed is the state to mount in. jumpToPage reloads the whole
// app on a transport without onNavigate, so this is what carries a collapsed
// column across a page change; read it before the first paint, or the column
// shows up expanded and then jumps.
export function initialNavCollapsed(): boolean {
  return storedNavCollapsed() ?? DEFAULT_COLLAPSED
}

export function storeNavCollapsed(collapsed: boolean) {
  setStoredValue(STORAGE_KEY, collapsed ? 'true' : 'false')
}

const WIDTH_KEY = 'sidenav_width'

// The width the column has always had, and what a reset goes back to.
export const NAV_DEFAULT_WIDTH = 240

// Narrow enough to hand the page real room, wide enough for a page list to
// stay readable. Past the upper bound the column stops being a side column.
export const NAV_MIN_WIDTH = 180
export const NAV_MAX_WIDTH = 480

// The step an arrow key moves. Home and End go straight to the bounds.
export const NAV_WIDTH_STEP = 16

export function clampNavWidth(px: number): number {
  return Math.min(NAV_MAX_WIDTH, Math.max(NAV_MIN_WIDTH, Math.round(px)))
}

// initialNavWidth is the width to mount in, read before the first paint for
// the same reason initialNavCollapsed is. Number('') is 0 and
// Number(undefined) is NaN, so an unset or corrupt value falls back here
// rather than mounting a zero-width column.
export function initialNavWidth(): number {
  const stored = Number(getStoredValue(WIDTH_KEY))
  if (!Number.isFinite(stored) || stored <= 0) {
    return NAV_DEFAULT_WIDTH
  }

  return clampNavWidth(stored)
}

export function storeNavWidth(px: number) {
  setStoredValue(WIDTH_KEY, String(px))
}
