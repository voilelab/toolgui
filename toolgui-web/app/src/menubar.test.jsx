import React from 'react'
import { cleanup, fireEvent, render, screen, waitFor } from '@testing-library/react'
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
  {
    type: 'submenu',
    label: 'Help',
    children: [{ type: 'text', label: 'About', id: 'menu_item_about' }],
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

  // A menubar is armed by a click. A pointer crossing the row on its way
  // somewhere else must not pop a dropdown open.
  test('hovering opens nothing while the row is closed', async () => {
    renderApp({}, { menu: MENU })

    fireEvent.mouseOver(screen.getByRole('button', { name: 'File' }))
    await waitFor(() => {
      expect(screen.queryByText('Open')).not.toBeInTheDocument()
    })
  })

  // Once one is open, moving along the row moves the dropdown with the
  // pointer -- and takes the old one down, rather than leaving two up.
  test('only one entry is open at a time', async () => {
    renderApp({}, { menu: MENU })

    screen.getByRole('button', { name: 'File' }).click()
    await screen.findByText('Open')

    fireEvent.mouseOver(screen.getByRole('button', { name: 'Help' }))
    await screen.findByText('About')
    await waitFor(() => {
      expect(screen.queryByText('Open')).not.toBeInTheDocument()
    })
  })

  // A plain entry has no dropdown to move to, so crossing it closes what was
  // open rather than leaving a dropdown belonging to an entry left behind.
  test('crossing a plain entry closes what was open', async () => {
    renderApp({}, { menu: MENU })

    screen.getByRole('button', { name: 'File' }).click()
    await screen.findByText('Open')

    fireEvent.mouseOver(screen.getByRole('button', { name: 'Run' }))
    await waitFor(() => {
      expect(screen.queryByText('Open')).not.toBeInTheDocument()
    })
  })

  test('picking an item closes the dropdown', async () => {
    renderApp({}, { menu: MENU })

    screen.getByRole('button', { name: 'File' }).click()
    const open = await screen.findByText('Open')

    open.click()
    await waitFor(() => {
      expect(screen.queryByText('Open')).not.toBeInTheDocument()
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
