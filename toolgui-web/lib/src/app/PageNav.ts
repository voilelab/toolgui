import { createContext, useContext } from "react"
import { PageQuery, pageHref, withStickyQuery } from "./pageurl"

// PageNav is how a component links to another page of the app.
export interface PageNav {
  href(name: string, query: PageQuery): string
  navigate(name: string, query: PageQuery): void
}

// newPageNav navigates with onNavigate when the transport routes itself (wasm,
// a desktop webview), else by moving the browser. Every link carries the
// sticky keys of current, the page query being read.
export function newPageNav(
  hashMode: boolean,
  onNavigate?: (name: string, query: string) => void,
  current: string = '',
  stickyKeys: string[] = []): PageNav {

  // A transport with its own navigation has no URL to point at, so it gets
  // the hash form.
  const useHash = hashMode || !!onNavigate
  const sticky = (query: PageQuery) => withStickyQuery(query, current, stickyKeys)

  return {
    href: (name, query) => pageHref(name, sticky(query), useHash),
    navigate: (name, query) => {
      if (onNavigate) {
        onNavigate(name, sticky(query))
        return
      }

      window.location.href = pageHref(name, sticky(query), hashMode)

      // A hash change alone opens no new session on the web.
      if (hashMode) {
        window.location.reload()
      }
    },
  }
}

export const PageNavContext = createContext<PageNav>(newPageNav(false))

export function usePageNav(): PageNav {
  return useContext(PageNavContext)
}
