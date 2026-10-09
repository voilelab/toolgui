import React from 'react'
import { act, cleanup, render, screen } from '@testing-library/react'
import { afterEach, expect, test, describe, vi } from 'vitest'

import { App, dispatchPack } from '@toolgui-web/lib'

function appConf(sticky) {
  return {
    page_names: ['index', 'detail'],
    page_confs: {
      index: { name: 'index', title: 'Index', emoji: '' },
      detail: { name: 'detail', title: 'Detail', emoji: '' },
    },
    title: '',
    main_container_id: 'container_main',
    sidebar_container_id: 'container_sidebar',
    hash_page_name_mode: false,
    version: 'v0.0.0',
    show_version: false,
    sticky_query: sticky,
  }
}

const RENDER_PROPS = { update: vi.fn(), upload: vi.fn(), download: vi.fn() }

function renderApp(props) {
  const ref = React.createRef()
  render(<App ref={ref} {...RENDER_PROPS} {...props} />)
  return ref.current
}

// drawPageLink runs the page once, drawing a PageLink to detail.
function drawPageLink(app) {
  act(() => {
    dispatchPack(app, { ready: true })
    dispatchPack(app, {
      type: 1, key: 'page_link_0', parent_key: 'container_main', index: 0,
      component: {
        name: 'page_link_component', id: '', text: 'Open', page: 'detail',
        query: { id: ['7'] },
      },
    })
    dispatchPack(app, { success: true })
  })
}

const navHref = () =>
  screen.getByRole('navigation', { name: 'main navigation' })
    .querySelector('a[href^="/detail"]').getAttribute('href')
const linkHref = () => screen.getByText('Open').closest('a').getAttribute('href')

afterEach(() => {
  cleanup()
  window.history.replaceState(null, '', '/')
})

describe('sticky query', () => {
  test('links carry the sticky keys, and follow ReplaceQuery', () => {
    window.history.replaceState(null, '', '/index?group=a&x=1')

    const app = renderApp({ appConf: appConf(['group']) })
    drawPageLink(app)

    expect(navHref()).toBe('/detail?group=a')
    expect(linkHref()).toBe('/detail?id=7&group=a')

    act(() => { dispatchPack(app, { replace_query: { group: ['b'] } }) })

    expect(navHref()).toBe('/detail?group=b')
    expect(linkHref()).toBe('/detail?id=7&group=b')
  })

  test('the transport query seeds the links', () => {
    const app = renderApp({
      appConf: appConf(['group']), pageName: 'index', query: 'group=a',
      onNavigate: vi.fn(), onReplaceQuery: vi.fn(),
    })
    drawPageLink(app)

    expect(linkHref()).toBe('#/detail?id=7&group=a')
  })

  test('links are unchanged without sticky keys', () => {
    window.history.replaceState(null, '', '/index?group=a')

    const app = renderApp({ appConf: appConf(undefined) })
    drawPageLink(app)

    expect(navHref()).toBe('/detail')
    expect(linkHref()).toBe('/detail?id=7')
  })
})
