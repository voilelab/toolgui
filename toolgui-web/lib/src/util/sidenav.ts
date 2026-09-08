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
