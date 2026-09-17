import React from 'react'
import { act, cleanup, fireEvent, render, screen } from '@testing-library/react'
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

// The drag handle. Its aria-valuenow is what the width is readable from --
// the width itself lands in a CSS variable jsdom does not compute.
function resizer() {
  return screen.getByRole('separator', { name: /Resize the side column/ })
}

function navWidthVar() {
  return document.querySelector('.toolgui-nav').style.getPropertyValue('--tg-nav-width')
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

describe('side column width', () => {
  test('starts at the default for a visitor who never dragged', () => {
    renderApp()
    expect(resizer()).toHaveAttribute('aria-valuenow', '240')
    expect(navWidthVar()).toBe('240px')
  })

  test('the handle reports the bounds it moves between', () => {
    renderApp()
    expect(resizer()).toHaveAttribute('aria-valuemin', '180')
    expect(resizer()).toHaveAttribute('aria-valuemax', '480')
  })

  test('arrow keys step the width and store it', () => {
    renderApp()
    fireEvent.keyDown(resizer(), { key: 'ArrowRight' })

    expect(resizer()).toHaveAttribute('aria-valuenow', '256')
    expect(navWidthVar()).toBe('256px')
    expect(window.localStorage.getItem('sidenav_width')).toBe('256')

    fireEvent.keyDown(resizer(), { key: 'ArrowLeft' })
    expect(resizer()).toHaveAttribute('aria-valuenow', '240')
  })

  test('Home and End go to the bounds, and stop there', () => {
    renderApp()
    fireEvent.keyDown(resizer(), { key: 'Home' })
    expect(resizer()).toHaveAttribute('aria-valuenow', '180')

    fireEvent.keyDown(resizer(), { key: 'ArrowLeft' })
    expect(resizer()).toHaveAttribute('aria-valuenow', '180')

    fireEvent.keyDown(resizer(), { key: 'End' })
    expect(resizer()).toHaveAttribute('aria-valuenow', '480')

    fireEvent.keyDown(resizer(), { key: 'ArrowRight' })
    expect(resizer()).toHaveAttribute('aria-valuenow', '480')
  })

  test('Enter and a double click both reset the width', () => {
    renderApp()
    fireEvent.keyDown(resizer(), { key: 'End' })

    fireEvent.keyDown(resizer(), { key: 'Enter' })
    expect(resizer()).toHaveAttribute('aria-valuenow', '240')

    fireEvent.keyDown(resizer(), { key: 'End' })
    fireEvent.doubleClick(resizer())
    expect(resizer()).toHaveAttribute('aria-valuenow', '240')
    expect(window.localStorage.getItem('sidenav_width')).toBe('240')
  })

  // Same as the collapsed state: jumpToPage reloads the app, so only what
  // was stored survives a page change.
  test('a fresh mount comes up at the stored width', () => {
    window.localStorage.setItem('sidenav_width', '320')
    renderApp()

    expect(resizer()).toHaveAttribute('aria-valuenow', '320')
    expect(navWidthVar()).toBe('320px')
  })

  // Number('') is 0 and Number(undefined) is NaN; neither may mount a
  // zero-width column.
  test.each(['', 'wide', '0', '-40'])('ignores a stored %s', (stored) => {
    window.localStorage.setItem('sidenav_width', stored)
    renderApp()
    expect(resizer()).toHaveAttribute('aria-valuenow', '240')
  })

  test('clamps a stored width from outside the bounds', () => {
    window.localStorage.setItem('sidenav_width', '2000')
    renderApp()
    expect(resizer()).toHaveAttribute('aria-valuenow', '480')
  })

  test('renders at the default when storage is unavailable', () => {
    vi.spyOn(window.localStorage, 'getItem').mockImplementation(() => {
      throw new Error('opaque origin')
    })
    vi.spyOn(window.localStorage, 'setItem').mockImplementation(() => {
      throw new Error('opaque origin')
    })

    renderApp()
    expect(resizer()).toHaveAttribute('aria-valuenow', '240')

    fireEvent.keyDown(resizer(), { key: 'ArrowRight' })
    expect(resizer()).toHaveAttribute('aria-valuenow', '256')
  })

  test('a drag moves the edge and stores where it landed', () => {
    renderApp()
    const handle = resizer()

    fireEvent.pointerDown(handle, { pointerId: 1, clientX: 240 })
    fireEvent.pointerMove(handle, { pointerId: 1, clientX: 300 })
    fireEvent.pointerUp(handle, { pointerId: 1, clientX: 300 })

    expect(resizer()).toHaveAttribute('aria-valuenow', '300')
    expect(navWidthVar()).toBe('300px')
    expect(window.localStorage.getItem('sidenav_width')).toBe('300')
  })

  // fireEvent flushes between calls, which a browser dispatching all three in
  // one task does not. Raw dispatches inside a single act() are that case: a
  // drag that read its own flag off state would see it unset and drop every
  // move, leaving the edge where it was and the body class behind.
  test('a drag dispatched in one tick still moves the edge', () => {
    renderApp()
    const handle = resizer()
    const send = (type, clientX) => handle.dispatchEvent(
      new MouseEvent(type, { bubbles: true, cancelable: true, clientX }))

    act(() => {
      send('pointerdown', 240)
      send('pointermove', 300)
      send('pointerup', 300)
    })

    expect(resizer()).toHaveAttribute('aria-valuenow', '300')
    expect(document.body).not.toHaveClass('toolgui-resizing')
  })

  // A right press opens the context menu. Dragging under it must not take the
  // edge along, nor store where it ended up.
  test('a non-primary button does not start a drag', () => {
    renderApp()
    const handle = resizer()

    fireEvent.pointerDown(handle, { pointerId: 1, clientX: 240, button: 2 })
    fireEvent.pointerMove(handle, { pointerId: 1, clientX: 400 })
    fireEvent.pointerUp(handle, { pointerId: 1, clientX: 400 })

    expect(resizer()).toHaveAttribute('aria-valuenow', '240')
    expect(document.body).not.toHaveClass('toolgui-resizing')
    expect(window.localStorage.getItem('sidenav_width')).toBe(null)
  })

  // A second finger landing on the handle must not move an edge the first one
  // is already holding.
  test('a second pointer does not join a drag in flight', () => {
    renderApp()
    const handle = resizer()

    fireEvent.pointerDown(handle, { pointerId: 1, clientX: 240, button: 0 })
    fireEvent.pointerMove(handle, { pointerId: 2, clientX: 460 })
    expect(resizer()).toHaveAttribute('aria-valuenow', '240')

    fireEvent.pointerUp(handle, { pointerId: 2, clientX: 460 })
    expect(document.body).toHaveClass('toolgui-resizing')

    // The one that started it still finishes it.
    fireEvent.pointerMove(handle, { pointerId: 1, clientX: 300 })
    fireEvent.pointerUp(handle, { pointerId: 1, clientX: 300 })
    expect(resizer()).toHaveAttribute('aria-valuenow', '300')
    expect(document.body).not.toHaveClass('toolgui-resizing')
  })

  // The class that stops the drag selecting the page it sweeps over must not
  // outlive the drag, or the whole document stays unselectable.
  test('a drag cleans up after itself', () => {
    renderApp()
    const handle = resizer()

    fireEvent.pointerDown(handle, { pointerId: 1, clientX: 240 })
    expect(document.body).toHaveClass('toolgui-resizing')

    fireEvent.pointerUp(handle, { pointerId: 1, clientX: 240 })
    expect(document.body).not.toHaveClass('toolgui-resizing')
  })

  test('a cancelled drag keeps where it got to, and cleans up', () => {
    renderApp()
    const handle = resizer()

    fireEvent.pointerDown(handle, { pointerId: 1, clientX: 240 })
    fireEvent.pointerMove(handle, { pointerId: 1, clientX: 280 })
    fireEvent.pointerCancel(handle, { pointerId: 1, clientX: 280 })

    expect(resizer()).toHaveAttribute('aria-valuenow', '280')
    expect(document.body).not.toHaveClass('toolgui-resizing')
  })

  // An unmount mid-drag would otherwise leave the document unselectable.
  test('unmounting mid-drag cleans up too', () => {
    renderApp()
    fireEvent.pointerDown(resizer(), { pointerId: 1, clientX: 240 })
    cleanup()

    expect(document.body).not.toHaveClass('toolgui-resizing')
  })

  test('a move with no drag under it does nothing', () => {
    renderApp()
    fireEvent.pointerMove(resizer(), { pointerId: 1, clientX: 400 })
    expect(resizer()).toHaveAttribute('aria-valuenow', '240')
  })

  // A key the separator has no business acting on has to reach the page.
  test('leaves other keys alone', () => {
    renderApp()
    const event = new KeyboardEvent('keydown',
      { key: 'a', bubbles: true, cancelable: true })
    resizer().dispatchEvent(event)

    expect(event.defaultPrevented).toBe(false)
    expect(resizer()).toHaveAttribute('aria-valuenow', '240')
  })
})
