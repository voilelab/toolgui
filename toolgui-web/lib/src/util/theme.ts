import { MantineColorSchemeManager } from '@mantine/core'

import { clearStoredValue, getStoredValue, setStoredValue } from './storage'

// The two themes the stylesheets know about. A visitor who has never touched
// the toggle has no stored preference, so the theme has to come from
// somewhere else than storage.
export type ThemeMode = 'light' | 'dark'

const STORAGE_KEY = 'theme_mode'

// preferredThemeMode asks the browser what the visitor wants, and is the
// theme to start in until the toggle stores a choice. matchMedia is guarded
// because a webview old enough to lack it should still get a theme.
export function preferredThemeMode(): ThemeMode {
  if (typeof window.matchMedia !== 'function') {
    return 'light'
  }

  return window.matchMedia('(prefers-color-scheme: dark)').matches ?
    'dark' : 'light'
}

// themeModeManager lets Mantine hold the color scheme, the app's one source
// of truth for the theme, while the preference stays under the app's own
// storage key.
export const themeModeManager: MantineColorSchemeManager = {
  get: (defaultValue) => {
    const stored = getStoredValue(STORAGE_KEY)
    if (stored === 'light' || stored === 'dark') {
      return stored
    }

    return defaultValue
  },

  // 'auto' never gets here: the toggle only ever sets a concrete mode.
  set: (value) => {
    if (value === 'light' || value === 'dark') {
      setStoredValue(STORAGE_KEY, value)
    }
  },

  // Nothing outside Mantine changes the theme, so there is nothing to
  // subscribe to.
  subscribe: () => { },
  unsubscribe: () => { },

  clear: () => { clearStoredValue(STORAGE_KEY) },
}

// applyBulmaTheme puts the theme on <html>, where Bulma's dark scheme picks
// it up. The app's own CSS reads Mantine's variables instead, so this goes
// away with Bulma.
export function applyBulmaTheme(mode: ThemeMode) {
  document.documentElement.className = 'theme-' + mode
}
