import React from 'react'
import { cleanup, render } from '@testing-library/react'
import { afterEach, expect, test, describe, vi } from 'vitest'

import { App, dispatchPack } from '@toolgui-web/lib'

function appConf(hashMode) {
  return {
    page_names: ['index', 'detail'],
    page_confs: {
      index: { name: 'index', title: 'Index', emoji: '' },
      detail: { name: 'detail', title: 'Detail', emoji: '' },
    },
    title: '',
    main_container_id: 'container_main',
    sidebar_container_id: 'container_sidebar',
    hash_page_name_mode: hashMode,
    version: 'v0.0.0',
    show_version: false,
  }
}

const RENDER_PROPS = { update: vi.fn(), upload: vi.fn(), download: vi.fn() }

function renderApp(props) {
  const ref = React.createRef()
  render(<App ref={ref} {...RENDER_PROPS} {...props} />)
  return ref.current
}

afterEach(() => {
  cleanup()
  window.history.replaceState(null, '', '/')
})

describe('replace query', () => {
  test('replaces the path mode query without a history entry', () => {
    window.history.replaceState(null, '', '/detail?group=a')
    const before = window.history.length

    const app = renderApp({ appConf: appConf(false) })
    dispatchPack(app, { replace_query: { group: ['b c'], x: ['1&2'] } })

    expect(window.location.pathname).toBe('/detail')
    expect(window.location.search).toBe('?group=b+c&x=1%262')
    expect(window.history.length).toBe(before)
  })

  test('keeps the real query string in hash mode', () => {
    window.history.replaceState(null, '', '/?embed=1#/detail?group=a')

    const app = renderApp({ appConf: appConf(true) })
    dispatchPack(app, { replace_query: { group: ['b'] } })

    expect(window.location.search).toBe('?embed=1')
    expect(window.location.hash).toBe('#/detail?group=b')
  })

  test('an empty query drops the ?', () => {
    window.history.replaceState(null, '', '/#/detail?group=a')

    const app = renderApp({ appConf: appConf(true) })
    dispatchPack(app, { replace_query: {} })

    expect(window.location.hash).toBe('#/detail')
  })

  test('goes to onReplaceQuery when the transport has one', () => {
    window.history.replaceState(null, '', '/')
    const onReplaceQuery = vi.fn()

    const app = renderApp({
      appConf: appConf(false), pageName: 'detail', onReplaceQuery,
    })
    dispatchPack(app, { replace_query: { group: ['b'] } })

    expect(onReplaceQuery).toHaveBeenCalledWith('group=b')
    expect(window.location.href).toBe(window.location.origin + '/')
  })
})
