import React from 'react'
import { cleanup, fireEvent } from '@testing-library/react'

import { render } from './render'
import { afterEach, expect, test, describe, vi } from 'vitest'

import { Node } from '@toolgui-web/lib/src/app/Nodes'
import { TPageLink } from '@toolgui-web/lib/src/components/tccontent/page_link'
import { PageNavContext, newPageNav } from '@toolgui-web/lib/src/app/PageNav'

const RENDER_PROPS = { update: vi.fn(), upload: vi.fn(), theme: 'light' }

afterEach(cleanup)

const link = (nav) => render(
  <PageNavContext.Provider value={nav}>
    <TPageLink node={new Node('main/0', {
      name: 'page_link_component', id: '', text: 'Open', page: 'detail',
      query: { group: ['a b'] },
    })} {...RENDER_PROPS} />
  </PageNavContext.Provider>
)

describe('page link', () => {
  test('is a real link in path mode', () => {
    const { container } = link(newPageNav(false))
    expect(container.querySelector('a').getAttribute('href'))
      .toBe('/detail?group=a+b')
  })

  test('is a hash link in hash mode', () => {
    const { container } = link(newPageNav(true))
    expect(container.querySelector('a').getAttribute('href'))
      .toBe('#/detail?group=a+b')
  })

  test('a click goes through onNavigate with the query', () => {
    const onNavigate = vi.fn()
    const { container } = link(newPageNav(false, onNavigate))
    const a = container.querySelector('a')

    expect(a.getAttribute('href')).toBe('#/detail?group=a+b')
    fireEvent.click(a)
    expect(onNavigate).toHaveBeenCalledWith('detail', 'group=a+b')
  })

  test('a modified click is the browser\'s', () => {
    const onNavigate = vi.fn()
    const { container } = link(newPageNav(false, onNavigate))

    fireEvent.click(container.querySelector('a'), { ctrlKey: true })
    expect(onNavigate).not.toHaveBeenCalled()
  })
})
