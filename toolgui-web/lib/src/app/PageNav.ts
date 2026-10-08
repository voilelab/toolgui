import { createContext, useContext } from "react"
import { PageQuery, encodeQuery, pageHref } from "./pageurl"

// PageNav is how a component links to another page of the app.
export interface PageNav {
  href(name: string, query: PageQuery): string
  navigate(name: string, query: PageQuery): void
}

// newPageNav navigates with onNavigate when the transport routes itself (wasm,
// a desktop webview), else by moving the browser.
export function newPageNav(
  hashMode: boolean,
  onNavigate?: (name: string, query: string) => void): PageNav {

  // A transport with its own navigation has no URL to point at, so it gets
  // the hash form.
  const useHash = hashMode || !!onNavigate

  return {
    href: (name, query) => pageHref(name, query, useHash),
    navigate: (name, query) => {
      if (onNavigate) {
        onNavigate(name, encodeQuery(query))
        return
      }

      window.location.href = pageHref(name, query, hashMode)

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
