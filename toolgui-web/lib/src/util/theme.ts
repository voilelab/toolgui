import { getStoredValue, setStoredValue } from './storage'

// The two themes the stylesheets know about. A visitor who has never touched
// the toggle has no stored preference, so the theme has to come from
// somewhere else than storage.
export type ThemeMode = 'light' | 'dark'

const STORAGE_KEY = 'darkMode'

// preferredThemeMode asks the browser what the visitor wants. matchMedia is
// guarded because a webview old enough to lack it should still get a theme.
function preferredThemeMode(): ThemeMode {
  if (typeof window.matchMedia !== 'function') {
    return 'light'
  }

  return window.matchMedia('(prefers-color-scheme: dark)').matches ?
    'dark' : 'light'
}

// initialThemeMode is the theme to start in: the stored choice once the
// visitor has made one, and the browser's preference until then.
export function initialThemeMode(): ThemeMode {
  const stored = getStoredValue(STORAGE_KEY)
  if (stored === 'light' || stored === 'dark') {
    return stored
  }

  return preferredThemeMode()
}

export function storeThemeMode(mode: ThemeMode) {
  setStoredValue(STORAGE_KEY, mode)
}

// applyThemeMode puts the theme on <html>, where Bulma and the app's own CSS
// pick it up.
export function applyThemeMode(mode: ThemeMode) {
  document.documentElement.className = 'theme-' + mode
}
