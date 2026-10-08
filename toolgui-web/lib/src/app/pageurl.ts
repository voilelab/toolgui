// Page URLs. The page query always follows the page name, wherever the name
// is: `/detail?group=a` in path mode, `#/detail?group=a` in hash mode and wasm.
// The real query string stays the app's (`?embed`, tgwasm.Query).

// PageQuery is a page query: encoded (`group=a`), or the url.Values a
// component carries.
export type PageQuery = string | Record<string, string[]> | null | undefined

// PageLocation is a page name and its encoded query, without the `?`.
export interface PageLocation {
  name: string
  query: string
}

// splitPagePart splits the page part of a URL -- the path or the hash after
// `/` -- at the first `?`.
export function splitPagePart(part: string): PageLocation {
  const i = part.indexOf('?')
  const name = i < 0 ? part : part.substring(0, i)
  const query = i < 0 ? '' : part.substring(i + 1)

  return { name: decodeName(name), query }
}

// A browser may hand the name back percent-encoded; a broken escape is kept
// as it is.
function decodeName(name: string): string {
  try {
    return decodeURIComponent(name)
  } catch {
    return name
  }
}

// pageFromLocation reads the page off a location. In hash mode an empty hash
// is the first page.
export function pageFromLocation(
  loc: { pathname: string, search: string, hash: string },
  hashMode: boolean, pageNames: string[]): PageLocation {

  if (!hashMode) {
    const { name } = splitPagePart(loc.pathname.substring(1))
    return { name, query: loc.search.replace(/^\?/, '') }
  }

  if (loc.hash.startsWith('#/')) {
    return splitPagePart(loc.hash.substring(2))
  }

  return { name: pageNames.length > 0 ? pageNames[0] : '', query: '' }
}

// encodeQuery encodes a page query. Values go through URLSearchParams, so
// nothing in them can end the query or change where the URL points.
export function encodeQuery(query: PageQuery): string {
  if (!query) {
    return ''
  }

  if (typeof query === 'string') {
    return new URLSearchParams(query).toString()
  }

  const params = new URLSearchParams()
  for (const key of Object.keys(query)) {
    for (const value of query[key] || []) {
      params.append(key, value)
    }
  }

  return params.toString()
}

// pageHref builds the href of a page of this app. Hash mode keeps the current
// query string, so `?embed=1` stays.
export function pageHref(name: string, query: PageQuery, hashMode: boolean): string {
  const q = encodeQuery(query)
  const part = encodeURIComponent(name) + (q ? '?' + q : '')

  return (hashMode ? '#/' : '/') + part
}
