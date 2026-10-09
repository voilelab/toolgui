import React from 'react'
import { cleanup, render } from '@testing-library/react'
import { afterEach, expect, test, describe, vi } from 'vitest'

import { App, dispatchPack } from '@toolgui-web/lib'

const APP_CONF = {
  page_names: ['problems', 'detail'],
  page_confs: {
    problems: { name: 'problems', title: 'Problems', emoji: '' },
    detail: { name: 'detail', title: 'Detail', emoji: '' },
  },
  title: '',
  main_container_id: 'container_main',
  sidebar_container_id: 'container_sidebar',
  hash_page_name_mode: true,
  version: 'v0.0.0',
  show_version: false,
}

const RENDER_PROPS = { update: vi.fn(), upload: vi.fn(), download: vi.fn() }

function renderApp(onNavigate) {
  const ref = React.createRef()
  render(<App ref={ref} {...RENDER_PROPS} appConf={APP_CONF}
    pageName="problems" onNavigate={onNavigate} />)
  return ref.current
}

afterEach(cleanup)

describe('navigate', () => {
  test('goes through onNavigate with the query', () => {
    const onNavigate = vi.fn()
    const app = renderApp(onNavigate)

    dispatchPack(app, { navigate: { page: 'detail', query: { id: ['0004'] } } })
    expect(onNavigate).toHaveBeenCalledWith('detail', 'id=0004')
  })

  test('an unknown page goes nowhere', () => {
    const onNavigate = vi.fn()
    const app = renderApp(onNavigate)
    const error = vi.spyOn(console, 'error').mockImplementation(() => {})

    dispatchPack(app, { navigate: { page: '//evil.example', query: {} } })
    expect(onNavigate).not.toHaveBeenCalled()
    error.mockRestore()
  })
})
