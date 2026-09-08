import React from 'react'
import { cleanup, fireEvent, render, screen } from '@testing-library/react'
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
  version: 'v0.0.0',
  show_version: false,
}

const RENDER_PROPS = { update: vi.fn(), upload: vi.fn() }

function renderApp() {
  render(<App appConf={APP_CONF} pageName="index" {...RENDER_PROPS} />)
}

// The one control the column's collapsed state is readable from: its label
// and aria-expanded both flip with it.
function collapseToggle() {
  return screen.getByRole('button', { name: /the side column/ })
}

beforeEach(() => {
  window.localStorage.clear()
})

afterEach(() => {
  cleanup()
  vi.restoreAllMocks()
})

describe('side column collapse', () => {
  test('starts expanded for a visitor who never touched the toggle', () => {
    renderApp()
    expect(collapseToggle()).toHaveAttribute('aria-expanded', 'true')
  })

  test('collapsing stores the choice', () => {
    renderApp()
    fireEvent.click(collapseToggle())

    expect(collapseToggle()).toHaveAttribute('aria-expanded', 'false')
    expect(window.localStorage.getItem('sidenav_collapsed')).toBe('true')
  })

  // What a page change comes back to: jumpToPage reloads the whole app, so a
  // second mount is all that is left of the first one's state.
  test('a fresh mount comes up collapsed', () => {
    renderApp()
    fireEvent.click(collapseToggle())
    cleanup()

    renderApp()
    expect(collapseToggle()).toHaveAttribute('aria-expanded', 'false')
    expect(document.querySelector('.toolgui-nav')).toHaveClass('is-collapsed')
  })

  test('expanding again stores that too', () => {
    window.localStorage.setItem('sidenav_collapsed', 'true')
    renderApp()
    fireEvent.click(collapseToggle())

    expect(collapseToggle()).toHaveAttribute('aria-expanded', 'true')
    expect(window.localStorage.getItem('sidenav_collapsed')).toBe('false')
  })

  // A desktop webview loads the app from an opaque origin, where every
  // localStorage call throws.
  test('renders expanded when storage is unavailable', () => {
    vi.spyOn(window.localStorage, 'getItem').mockImplementation(() => {
      throw new Error('opaque origin')
    })
    vi.spyOn(window.localStorage, 'setItem').mockImplementation(() => {
      throw new Error('opaque origin')
    })

    renderApp()
    expect(collapseToggle()).toHaveAttribute('aria-expanded', 'true')

    fireEvent.click(collapseToggle())
    expect(collapseToggle()).toHaveAttribute('aria-expanded', 'false')
  })

  test('ignores a stored value it does not understand', () => {
    window.localStorage.setItem('sidenav_collapsed', 'yes')
    renderApp()
    expect(collapseToggle()).toHaveAttribute('aria-expanded', 'true')
  })
})
