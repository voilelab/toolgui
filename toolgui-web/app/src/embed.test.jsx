import React from 'react'
import { cleanup, render, screen } from '@testing-library/react'
import { afterEach, beforeEach, expect, test, describe, vi } from 'vitest'

import { App } from '@toolgui-web/lib'

const APP_CONF = {
  page_names: ['index', 'other'],
  page_confs: {
    index: { name: 'index', title: 'Index', emoji: '' },
    other: { name: 'other', title: 'Other', emoji: '' },
  },
  title: 'My Tool',
  main_container_id: 'container_main',
  sidebar_container_id: 'container_sidebar',
  hash_page_name_mode: true,
  version: 'v1.2.3',
  // On, so a version line left behind in embed mode shows up here.
  show_version: true,
}

const RENDER_PROPS = { update: vi.fn(), upload: vi.fn() }

function renderApp(props) {
  render(<App appConf={APP_CONF} pageName="index" {...RENDER_PROPS} {...props} />)
}

beforeEach(() => {
  window.localStorage.clear()
})

afterEach(() => {
  cleanup()
  vi.restoreAllMocks()
})

describe('embed mode', () => {
  test('leaves the page and nothing else', () => {
    renderApp({ embed: true })

    expect(screen.queryByRole('navigation', { name: 'main navigation' }))
      .not.toBeInTheDocument()
    expect(screen.queryByRole('link', { name: /Other/ })).not.toBeInTheDocument()
    expect(screen.queryByText(/toolgui v1\.2\.3/)).not.toBeInTheDocument()
    expect(screen.queryByRole('button', { name: /the side column/ }))
      .not.toBeInTheDocument()
    expect(screen.queryByRole('button', { name: 'Rerun' })).not.toBeInTheDocument()

    // The page itself is still there, and the shell says which mode it is in
    // -- the CSS hands the width back to the page off that class.
    expect(document.querySelector('.toolgui-page')).toBeInTheDocument()
    expect(document.querySelector('.toolgui-shell')).toHaveClass('is-embed')
  })

  // Nothing in the frame navigates, so a click inside it cannot take the
  // hash -- and with it the history of the document holding the iframe --
  // anywhere.
  test('has nothing left that changes the page', () => {
    renderApp({ embed: true })
    expect(document.querySelectorAll('a[href^="#/"]')).toHaveLength(0)
  })

  // The column is the width in a narrow frame: there is no room for both.
  test('drops the column the page shares its width with', () => {
    renderApp({ embed: true })

    expect(document.querySelector('.toolgui-nav')).not.toBeInTheDocument()
    expect(document.querySelector('.toolgui-nav-resizer')).not.toBeInTheDocument()
  })

  // embed is a mode a URL asks for. Without it the app is the app it was.
  test('is off unless it is asked for', () => {
    renderApp()

    expect(screen.getByRole('navigation', { name: 'main navigation' }))
      .toBeInTheDocument()
    expect(screen.getByRole('link', { name: /Other/ })).toBeInTheDocument()
    expect(screen.getByText(/toolgui v1\.2\.3/)).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /the side column/ }))
      .toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Rerun' })).toBeInTheDocument()
    expect(document.querySelector('.toolgui-shell')).not.toHaveClass('is-embed')
  })

  test('embed={false} is the same as leaving it out', () => {
    renderApp({ embed: false })

    expect(screen.getByRole('navigation', { name: 'main navigation' }))
      .toBeInTheDocument()
    expect(document.querySelector('.toolgui-shell')).not.toHaveClass('is-embed')
  })
})
