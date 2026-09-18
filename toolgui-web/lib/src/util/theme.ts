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

// The listener subscribe put on the window, kept so unsubscribe takes off
// the same one. There is one manager, so there is one of these.
let onStorage: ((event: StorageEvent) => void) | undefined

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

  // Something outside this document can write the key: the book publishes
  // the demo app beside itself, so the two share an origin and with it
  // localStorage, and it mirrors its own theme into this key to keep the
  // demo in its iframe on the same theme it is (docs/demo-theme.js).
  //
  // A storage event is how that arrives. It fires in every other document of
  // the origin and never in the one that did the writing, so following it
  // cannot loop back on the toggle -- and a write of the value already
  // stored fires nothing at all, which is what keeps a reader's own choice
  // from being overwritten by a theme that has not changed.
  subscribe: (onUpdate) => {
    onStorage = (event: StorageEvent) => {
      if (event.storageArea !== window.localStorage) {
        return
      }

      // A null key is the whole store being cleared, which says nothing
      // about the theme: leave the app on the one it is showing.
      if (event.key !== STORAGE_KEY) {
        return
      }

      if (event.newValue === 'light' || event.newValue === 'dark') {
        onUpdate(event.newValue)
      }
    }

    window.addEventListener('storage', onStorage)
  },

  unsubscribe: () => {
    if (onStorage) {
      window.removeEventListener('storage', onStorage)
      onStorage = undefined
    }
  },

  clear: () => { clearStoredValue(STORAGE_KEY) },
}
