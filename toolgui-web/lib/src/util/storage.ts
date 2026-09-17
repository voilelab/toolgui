// Every executor serves the app from an ordinary origin, desktop included, so
// localStorage is normally there. A visitor who blocks site data, or a store
// out of quota, still makes it throw -- and reading or writing a preference is
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

export function clearStoredValue(key: string) {
  try {
    window.localStorage.removeItem(key)
  } catch (e) {
    // No storage here, so there was nothing stored to clear.
  }
}
