import { describe, expect, it } from 'vitest'
import { newPageNav } from './PageNav'

describe('newPageNav', () => {
  it('carries sticky keys in path mode', () => {
    const nav = newPageNav(false, undefined, 'group=a&x=1', ['group'])
    expect(nav.href('detail', '')).toBe('/detail?group=a')
    expect(nav.href('detail', { group: ['b'] })).toBe('/detail?group=b')
  })

  it('carries sticky keys in hash mode', () => {
    const nav = newPageNav(true, undefined, 'group=a', ['group'])
    expect(nav.href('detail', 'id=1')).toBe('#/detail?id=1&group=a')
  })

  it('hands sticky keys to onNavigate', () => {
    let got = ''
    const nav = newPageNav(false, (name, query) => { got = `${name}?${query}` },
      'group=a', ['group'])
    expect(nav.href('detail', '')).toBe('#/detail?group=a')
    nav.navigate('detail', '')
    expect(got).toBe('detail?group=a')
  })

  it('adds nothing without sticky keys', () => {
    const nav = newPageNav(false, undefined, 'group=a')
    expect(nav.href('detail', '')).toBe('/detail')
  })
})
