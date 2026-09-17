import { beforeEach, describe, expect, it } from 'vitest'

import { installExternalLinkHandler } from './external_link'

// The origin wails serves the app from. In a packaged build it is
// wails://wails; jsdom only has http, and what matters either way is that
// the app's own links resolve against it and an external one does not.
const pageOrigin = 'http://localhost:3000'

describe('installExternalLinkHandler', () => {
  let opened: string[]

  // clickLink renders one link, clicks the element the selector picks, and
  // reports whether the click was taken away from the webview.
  function clickLink(html: string, selector = 'a'): boolean {
    document.body.innerHTML = html
    const event = new MouseEvent('click', { bubbles: true, cancelable: true })
    document.body.querySelector(selector).dispatchEvent(event)
    return event.defaultPrevented
  }

  beforeEach(() => {
    opened = []
    document.body.innerHTML = ''
    installExternalLinkHandler(document, (url) => opened.push(url), pageOrigin)
  })

  it('hands an external link to the system browser', () => {
    expect(clickLink('<a href="https://example.com/doc">doc</a>')).toBe(true)
    expect(opened).toEqual(['https://example.com/doc'])
  })

  it('catches a click on what is inside the link', () => {
    clickLink('<a href="http://example.com/"><span>go</span></a>', 'span')

    expect(opened).toEqual(['http://example.com/'])
  })

  // A packaged build serves the page from wails://wails, where the anchor
  // would resolve this to wails://example.com/docs. Reading it as https is
  // what keeps the click off the webview there; asserting https rather than
  // the http jsdom's base would give proves the href was not resolved
  // against the origin.
  it('reads a scheme-relative link as https', () => {
    expect(clickLink('<a href="//example.com/docs">docs</a>')).toBe(true)
    expect(opened).toEqual(['https://example.com/docs'])
  })

  it('leaves a malformed scheme-relative link alone', () => {
    expect(clickLink('<a href="//">nowhere</a>')).toBe(false)
    expect(opened).toEqual([])
  })

  it('leaves an in-app page link alone', () => {
    // What AppSideNav renders, in both of its forms.
    expect(clickLink('<a href="/other">other</a>')).toBe(false)
    expect(clickLink('<a href="#/other">other</a>')).toBe(false)
    expect(opened).toEqual([])
  })

  it('leaves a non-web scheme alone', () => {
    expect(clickLink('<a href="file:///etc/hosts">hosts</a>')).toBe(false)
    expect(clickLink('<a href="mailto:a@example.com">mail</a>')).toBe(false)
    expect(opened).toEqual([])
  })

  it('leaves an anchor with no href alone', () => {
    expect(clickLink('<a>nowhere</a>')).toBe(false)
    expect(opened).toEqual([])
  })

  it('leaves a middle click to the webview', () => {
    document.body.innerHTML = '<a href="https://example.com/">doc</a>'
    document.body.querySelector('a').dispatchEvent(
      new MouseEvent('click', { bubbles: true, cancelable: true, button: 1 }))

    expect(opened).toEqual([])
  })
})
