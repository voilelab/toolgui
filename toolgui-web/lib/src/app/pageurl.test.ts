import { describe, expect, it } from 'vitest'
import {
  encodeQuery, pageFromLocation, pageHref, splitPagePart, withStickyQuery,
} from './pageurl'

const loc = (pathname: string, search: string, hash: string) =>
  ({ pathname, search, hash })

describe('splitPagePart', () => {
  it('splits at the first ?', () => {
    expect(splitPagePart('detail?group=a&x=b?c'))
      .toEqual({ name: 'detail', query: 'group=a&x=b?c' })
  })

  it('has an empty query without ?', () => {
    expect(splitPagePart('detail')).toEqual({ name: 'detail', query: '' })
  })

  it('decodes the name', () => {
    expect(splitPagePart('%E9%A0%81?x=1').name).toBe('頁')
    expect(splitPagePart('%zz').name).toBe('%zz')
  })
})

describe('pageFromLocation', () => {
  it('reads the path and search in path mode', () => {
    expect(pageFromLocation(loc('/detail', '?group=a', ''), false, ['index']))
      .toEqual({ name: 'detail', query: 'group=a' })
  })

  it('reads the hash in hash mode, leaving the search alone', () => {
    expect(pageFromLocation(loc('/', '?embed=1', '#/detail?x=1'), true, ['index']))
      .toEqual({ name: 'detail', query: 'x=1' })
  })

  it('falls back to the first page in hash mode', () => {
    expect(pageFromLocation(loc('/', '', ''), true, ['index']))
      .toEqual({ name: 'index', query: '' })
  })
})

describe('pageHref', () => {
  it('builds both forms', () => {
    const q = { group: ['a b'], name: ['x&y', 'z'] }
    expect(pageHref('detail', q, false)).toBe('/detail?group=a+b&name=x%26y&name=z')
    expect(pageHref('detail', q, true)).toBe('#/detail?group=a+b&name=x%26y&name=z')
  })

  it('drops an empty query', () => {
    expect(pageHref('detail', {}, false)).toBe('/detail')
    expect(pageHref('detail', '', true)).toBe('#/detail')
  })

  it('keeps a value from changing where the link points', () => {
    const href = pageHref('detail', { x: ['#/evil?//host'] }, true)
    expect(href).toBe('#/detail?x=%23%2Fevil%3F%2F%2Fhost')
  })

  it('round-trips through the parser', () => {
    const href = pageHref('頁', { group: ['a'] }, true)
    expect(pageFromLocation(loc('/', '', href), true, []))
      .toEqual({ name: '頁', query: 'group=a' })
  })
})

describe('encodeQuery', () => {
  it('normalizes a string', () => {
    expect(encodeQuery('a=1&b=x y')).toBe('a=1&b=x+y')
    expect(encodeQuery(null)).toBe('')
  })
})

describe('withStickyQuery', () => {
  it('copies only the sticky keys', () => {
    expect(withStickyQuery('', 'group=a&x=1', ['group'])).toBe('group=a')
  })

  it('keeps every value of a key', () => {
    expect(withStickyQuery(null, 'g=a&g=b', ['g'])).toBe('g=a&g=b')
  })

  it('lets the link win', () => {
    expect(withStickyQuery({ group: ['b'] }, 'group=a&day=1', ['group', 'day']))
      .toBe('group=b&day=1')
    expect(withStickyQuery('group=b', 'group=a', ['group'])).toBe('group=b')
  })

  it('counts a key the link sets empty as its own', () => {
    expect(withStickyQuery({ group: [] }, 'group=a', ['group'])).toBe('')
  })

  it('skips a key the current query lacks', () => {
    expect(withStickyQuery('id=1', 'x=2', ['group'])).toBe('id=1')
  })

  it('changes nothing without sticky keys', () => {
    expect(withStickyQuery({ id: ['1'] }, 'group=a', [])).toBe('id=1')
  })
})
