// A desktop window has no tabs and no back button, so a link that navigates
// the webview replaces the whole app: the page's state, and the way back to
// it, are both gone. External links belong in the system browser instead.

// externalLinkURL returns the url a click should hand to the system browser,
// or null when the click is the app's own business.
export function externalLinkURL(target: EventTarget | null, pageOrigin: string): string | null {
  const anchor = target instanceof Element ? target.closest('a') : null
  if (!anchor || !anchor.getAttribute('href')) {
    return null
  }

  // Web schemes only. file:, javascript: and the rest are not the system
  // browser's to open.
  if (anchor.protocol !== 'http:' && anchor.protocol !== 'https:') {
    return null
  }

  // Same origin is in-app navigation — the side nav's page links, which
  // resolve against wails' own origin.
  if (anchor.origin === pageOrigin) {
    return null
  }

  return anchor.href
}

// installExternalLinkHandler sends every external link on the page to
// openURL. It listens on the document rather than on the links themselves,
// so components keep rendering plain anchors: the url stays copyable, and
// nothing about the web build has to change.
export function installExternalLinkHandler(
  doc: Document,
  openURL: (url: string) => void,
  pageOrigin: string,
): void {
  // Capture, so the click arrives even where a component stops it from
  // bubbling. What keeps this off in-app links is the checks above, not the
  // phase.
  doc.addEventListener('click', (event: MouseEvent) => {
    // A secondary button does not navigate, and an already-cancelled click
    // has been claimed by someone else.
    if (event.defaultPrevented || event.button !== 0) {
      return
    }

    const url = externalLinkURL(event.target, pageOrigin)
    if (!url) {
      return
    }

    event.preventDefault()
    openURL(url)
  }, true)
}
