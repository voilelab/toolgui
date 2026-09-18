import React from 'react'
import { act, cleanup, render, screen } from '@testing-library/react'
import { afterEach, beforeEach, expect, test, describe, vi } from 'vitest'

import { App } from '@toolgui-web/lib'

const APP_CONF = {
  page_names: ['index'],
  page_confs: {
    index: { name: 'index', title: 'Index', emoji: '' },
  },
  title: 'My Tool',
  main_container_id: 'container_main',
  sidebar_container_id: 'container_sidebar',
  hash_page_name_mode: true,
  version: 'v1.2.3',
  show_version: false,
}

const RENDER_PROPS = { update: vi.fn(), upload: vi.fn() }

const THEME_KEY = 'theme_mode'

function renderApp(props) {
  render(<App appConf={APP_CONF} pageName="index" {...RENDER_PROPS} {...props} />)
}

// The toggle says which theme the app is in: it offers the other one. That is
// the app following, rather than Mantine holding a value nothing drew.
function themeMode() {
  if (screen.queryByRole('button', { name: 'Switch to the dark theme' })) {
    return 'light'
  }

  if (screen.queryByRole('button', { name: 'Switch to the light theme' })) {
    return 'dark'
  }

  return null
}

// writeFromOutside is another document of the same origin writing the key --
// the book, which publishes the demo app beside itself and puts it on the
// theme the book is being read in. The store changes and the event follows;
// jsdom raises no event of its own, and a real browser raises it in every
// document but the one that wrote.
function writeFromOutside(key, value) {
  const oldValue = window.localStorage.getItem(key)
  window.localStorage.setItem(key, value)

  act(() => {
    window.dispatchEvent(new StorageEvent('storage', {
      key: key,
      oldValue: oldValue,
      newValue: value,
      storageArea: window.localStorage,
    }))
  })
}

beforeEach(() => {
  window.localStorage.clear()
})

afterEach(() => {
  cleanup()
  vi.restoreAllMocks()
})

describe('the stored theme', () => {
  test('is where the app starts', () => {
    window.localStorage.setItem(THEME_KEY, 'dark')
    renderApp()

    expect(themeMode()).toBe('dark')
  })

  test('is what the toggle writes', async () => {
    renderApp()
    expect(themeMode()).toBe('light')

    await act(async () => {
      screen.getByRole('button', { name: 'Switch to the dark theme' }).click()
    })

    expect(themeMode()).toBe('dark')
    expect(window.localStorage.getItem(THEME_KEY)).toBe('dark')
  })
})

describe('a theme written from outside the document', () => {
  // What the book does when the reader switches it to navy or ayu: the frame
  // on the page goes dark with it, in the same action and without reloading.
  test('is followed by the app', () => {
    window.localStorage.setItem(THEME_KEY, 'light')
    renderApp()
    expect(themeMode()).toBe('light')

    writeFromOutside(THEME_KEY, 'dark')
    expect(themeMode()).toBe('dark')

    writeFromOutside(THEME_KEY, 'light')
    expect(themeMode()).toBe('light')
  })

  test('leaves the app alone when it is some other key', () => {
    renderApp()
    expect(themeMode()).toBe('light')

    writeFromOutside('mdbook-theme', 'navy')
    expect(themeMode()).toBe('light')
  })

  test('leaves the app alone when it is not a theme the app has', () => {
    renderApp()
    expect(themeMode()).toBe('light')

    writeFromOutside(THEME_KEY, 'ayu')
    expect(themeMode()).toBe('light')
  })

  // Clearing the store says nothing about the theme -- there is no theme to
  // go to -- so the app stays on the one it is showing.
  test('leaves the app alone when the whole store is cleared', () => {
    window.localStorage.setItem(THEME_KEY, 'dark')
    renderApp()
    expect(themeMode()).toBe('dark')

    act(() => {
      window.localStorage.clear()
      window.dispatchEvent(new StorageEvent('storage', {
        key: null,
        oldValue: null,
        newValue: null,
        storageArea: window.localStorage,
      }))
    })

    expect(themeMode()).toBe('dark')
  })

  test('stops arriving once the app is gone', () => {
    renderApp()
    cleanup()

    // Nothing is mounted to follow it; the listener coming off with the app
    // is what keeps this from throwing.
    expect(() => { writeFromOutside(THEME_KEY, 'dark') }).not.toThrow()
  })
})
