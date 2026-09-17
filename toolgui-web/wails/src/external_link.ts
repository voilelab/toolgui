// A desktop window has no tabs and no back button, so a link that navigates
// the webview replaces the whole app: the page's state, and the way back to
// it, are both gone. External links belong in the system browser instead.

// linkURL is where a link would take a browser on the web. That is what
// anchor.href already says, with one exception: a packaged build serves the
// page from wails://wails, so a scheme-relative href like //example.com/doc
// resolves against that and comes out as wails://example.com/doc. There is
// no such site to visit, and on the web the same href would have been
// http(s), so read those off the attribute as https instead.
function linkURL(anchor: HTMLAnchorElement): URL | null {
  const href = anchor.getAttribute('href')?.trim()
  if (!href) {
    return null
  }

  try {
    return new URL(href.startsWith('//') ? 'https:' + href : anchor.href)
  } catch {
    return null
  }
}

// externalLinkURL returns the url a click should hand to the system browser,
// or null when the click is the app's own business.
export function externalLinkURL(target: EventTarget | null, pageOrigin: string): string | null {
  const anchor = target instanceof Element ? target.closest('a') : null
  const url = anchor ? linkURL(anchor) : null
  if (!url) {
    return null
  }

  // Web schemes only. file:, javascript: and the rest are not the system
  // browser's to open.
  if (url.protocol !== 'http:' && url.protocol !== 'https:') {
    return null
  }

  // Same origin is in-app navigation — the side nav's page links, which
  // resolve against wails' own origin.
  if (url.origin === pageOrigin) {
    return null
  }

  return url.href
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
