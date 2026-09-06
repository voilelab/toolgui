// A desktop webview loads the app from a data: URL, and its opaque origin
// makes any localStorage access throw. Reading or writing a preference is
// never worth taking the app down with it.

export function getStoredValue(key: string): string | undefined {
  try {
    const value = window.localStorage.getItem(key)
    return value === null ? undefined : value
  } catch (e) {
    return undefined
  }
}

export function setStoredValue(key: string, value: string) {
  try {
    window.localStorage.setItem(key, value)
  } catch (e) {
    // No storage here, so the setting just does not stick.
  }
}
