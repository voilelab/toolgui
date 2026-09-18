import React from 'react'
import { cleanup, render, screen, waitFor } from '@testing-library/react'
import { afterEach, beforeEach, expect, test, describe, vi } from 'vitest'

import { App } from '@toolgui-web/lib'

const MENU = [
  {
    type: 'submenu',
    label: 'File',
    children: [
      { type: 'text', label: 'Open', id: 'menu_item_file_open' },
      { type: 'separator' },
      {
        type: 'submenu',
        label: 'Recent',
        children: [
          { type: 'text', label: 'notes.txt', id: 'menu_item_recent_notes' },
        ],
      },
    ],
  },
  { type: 'text', label: 'Run', id: 'menu_item_run' },
]

const APP_CONF = {
  page_names: ['index'],
  page_confs: { index: { name: 'index', title: 'Index', emoji: '' } },
  title: 'My Tool',
  main_container_id: 'container_main',
  sidebar_container_id: 'container_sidebar',
  hash_page_name_mode: true,
  version: 'v1.2.3',
  show_version: true,
}

function renderApp(props, conf) {
  const update = vi.fn()
  render(<App appConf={{ ...APP_CONF, ...conf }} pageName="index"
    update={update} upload={vi.fn()} {...props} />)

  return update
}

beforeEach(() => {
  window.localStorage.clear()
})

afterEach(() => {
  cleanup()
  vi.restoreAllMocks()
})

describe('menubar', () => {
  test('is a row above the shell', () => {
    renderApp({}, { menu: MENU })

    const bar = document.querySelector('.toolgui-menubar')
    expect(bar).toBeInTheDocument()
    expect(document.querySelector('.toolgui-frame'))
      .toHaveClass('has-menubar')

    // Above the shell, not inside it.
    expect(bar.nextElementSibling).toHaveClass('toolgui-shell')
    expect(document.querySelector('.toolgui-shell .toolgui-menubar'))
      .not.toBeInTheDocument()
  })

  // The whole point of the row being conditional: an app that declares no
  // menu must not pay for an empty one.
  test('is not in the document without a menu', () => {
    renderApp()

    expect(document.querySelector('.toolgui-menubar')).not.toBeInTheDocument()
    expect(document.querySelector('.toolgui-frame'))
      .not.toHaveClass('has-menubar')
  })

  test('an empty tree counts as no menu', () => {
    renderApp({}, { menu: [] })

    expect(document.querySelector('.toolgui-menubar')).not.toBeInTheDocument()
  })

  // Embed mode is the page on its own, and the menubar is the app's chrome.
  test('is dropped in embed mode', () => {
    renderApp({ embed: true }, { menu: MENU })

    expect(document.querySelector('.toolgui-menubar')).not.toBeInTheDocument()
  })

  test('a top level text item sends its click', async () => {
    const update = renderApp({}, { menu: MENU })

    screen.getByRole('button', { name: 'Run' }).click()

    expect(update).toHaveBeenCalledWith({
      type: 'click',
      id: 'menu_item_run',
    })
  })

  test('a submenu item sends its click', async () => {
    const update = renderApp({}, { menu: MENU })

    screen.getByRole('button', { name: 'File' }).click()
    const open = await screen.findByText('Open')

    open.click()
    await waitFor(() => {
      expect(update).toHaveBeenCalledWith({
        type: 'click',
        id: 'menu_item_file_open',
      })
    })
  })

  test('a submenu holds its separator and its own submenu', async () => {
    renderApp({}, { menu: MENU })

    screen.getByRole('button', { name: 'File' }).click()
    await screen.findByText('Open')

    expect(document.querySelectorAll('.mantine-Menu-divider'))
      .toHaveLength(1)
    expect(screen.getByText('Recent')).toBeInTheDocument()
  })
})
