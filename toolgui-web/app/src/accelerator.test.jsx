import React from 'react'
import { cleanup, fireEvent, render, screen } from '@testing-library/react'
import { afterEach, expect, test, describe, vi } from 'vitest'

import { App } from '@toolgui-web/lib'

// The browser's half of the feature: nothing dispatches an accelerator here,
// so the menubar listens for the keystroke itself. On the desktop the OS does
// it off the native menu item and none of this runs.

const MENU = [
  {
    type: 'submenu',
    label: 'File',
    children: [
      {
        type: 'text', label: 'Open', id: 'menu_item_file_open',
        accelerator: 'CmdOrCtrl+o',
      },
      {
        type: 'text', label: 'Save as', id: 'menu_item_save_as',
        accelerator: 'CmdOrCtrl+Shift+s',
      },
      { type: 'text', label: 'Quit', id: 'menu_item_file_quit' },
    ],
  },
  {
    type: 'text', label: 'Help', id: 'menu_item_help',
    accelerator: 'f2',
  },
  {
    type: 'text', label: 'Find', id: 'menu_item_find',
    accelerator: 'CmdOrCtrl+Shift+/',
  },
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

function renderApp(conf) {
  const update = vi.fn()
  const r = render(<App appConf={{ ...APP_CONF, ...conf }} pageName="index"
    update={update} upload={vi.fn()} />)

  return { update, ...r }
}

// The platform is read once, in the constructor, so this goes in before the
// render. `platform` is deprecated and read-only, hence the redefine.
function asMac(mac) {
  Object.defineProperty(window.navigator, 'platform', {
    value: mac ? 'MacIntel' : 'Linux x86_64',
    configurable: true,
  })
}

// press returns false where the keystroke was preventDefault'd, which is what
// dispatchEvent reports for a cancelled event.
function press(target, init) {
  return fireEvent.keyDown(target, { cancelable: true, ...init })
}

afterEach(() => {
  cleanup()
  asMac(false)
  vi.restoreAllMocks()
})

describe('menu accelerators', () => {
  test('a chord fires the item it was declared on', () => {
    const { update } = renderApp({ menu: MENU })

    expect(press(document.body, { key: 'o', code: 'KeyO', ctrlKey: true }))
      .toBe(false)

    expect(update).toHaveBeenCalledWith({
      type: 'click',
      id: 'menu_item_file_open',
    })
  })

  // A combination carrying a modifier the item did not ask for belongs to
  // whatever asked for that one, not to this item.
  test('an extra modifier is a different combination', () => {
    const { update } = renderApp({ menu: MENU })

    press(document.body,
      { key: 'O', code: 'KeyO', ctrlKey: true, shiftKey: true })

    expect(update).not.toHaveBeenCalled()
  })

  test('every declared modifier is required', () => {
    const { update } = renderApp({ menu: MENU })

    press(document.body, { key: 'S', code: 'KeyS', ctrlKey: true })
    expect(update).not.toHaveBeenCalled()

    press(document.body,
      { key: 'S', code: 'KeyS', ctrlKey: true, shiftKey: true })
    expect(update).toHaveBeenCalledWith({
      type: 'click',
      id: 'menu_item_save_as',
    })
  })

  // CmdOrCtrl is the platform's own menu modifier: Control here, Command on a
  // Mac. The same declaration, a different key.
  test('CmdOrCtrl is Control off a Mac', () => {
    const { update } = renderApp({ menu: MENU })

    press(document.body, { key: 'o', code: 'KeyO', metaKey: true })
    expect(update).not.toHaveBeenCalled()
  })

  test('CmdOrCtrl is Command on a Mac', () => {
    asMac(true)
    const { update } = renderApp({ menu: MENU })

    press(document.body, { key: 'o', code: 'KeyO', ctrlKey: true })
    expect(update).not.toHaveBeenCalled()

    press(document.body, { key: 'o', code: 'KeyO', metaKey: true })
    expect(update).toHaveBeenCalledWith({
      type: 'click',
      id: 'menu_item_file_open',
    })
  })

  test('an item with no accelerator is not reachable by keystroke', () => {
    const { update } = renderApp({ menu: MENU })

    press(document.body, { key: 'q', code: 'KeyQ', ctrlKey: true })
    expect(update).not.toHaveBeenCalled()
  })

  // The key an accelerator names is a physical key, not the character it
  // produces. Shift and `/` report `?`, so the character comparison is not
  // what can match it -- `code` is.
  test('a shifted punctuation key fires', () => {
    const { update } = renderApp({ menu: MENU })

    expect(press(document.body,
      { key: '?', code: 'Slash', ctrlKey: true, shiftKey: true })).toBe(false)

    expect(update).toHaveBeenCalledWith({ type: 'click', id: 'menu_item_find' })
  })

  // Another layout puts another character on the same key, and the item is
  // still the one that key fires.
  test('a layout that prints something else on the key still fires it', () => {
    const { update } = renderApp({ menu: MENU })

    press(document.body,
      { key: '-', code: 'Slash', ctrlKey: true, shiftKey: true })

    expect(update).toHaveBeenCalledWith({ type: 'click', id: 'menu_item_find' })
  })

  // An event with no code at all -- a synthetic one, or a key the browser
  // does not name -- matches nothing it did not match by character.
  test('an event with no code matches by character alone', () => {
    const { update } = renderApp({ menu: MENU })

    press(document.body, { key: '?', code: '', ctrlKey: true, shiftKey: true })
    expect(update).not.toHaveBeenCalled()
  })

  test('a named key fires too', () => {
    const { update } = renderApp({ menu: MENU })

    press(document.body, { key: 'F2', code: 'F2' })
    expect(update).toHaveBeenCalledWith({ type: 'click', id: 'menu_item_help' })
  })

  describe('while typing', () => {
    // The field is rendered outside the app: what the guard reads is the
    // event's target, not where in the tree it sits.
    function typeIn(tag, init) {
      const el = document.createElement(tag)
      document.body.appendChild(el)
      el.focus()

      press(el, init)

      return el
    }

    test('a bare key is a keystroke, not a shortcut', () => {
      const { update } = renderApp({ menu: MENU })

      typeIn('input', { key: 'F2', code: 'F2' })
      typeIn('textarea', { key: 'F2', code: 'F2' })

      expect(update).not.toHaveBeenCalled()
    })

    test('a chord still reaches the menu', () => {
      const { update } = renderApp({ menu: MENU })

      typeIn('input', { key: 'o', code: 'KeyO', ctrlKey: true })

      expect(update).toHaveBeenCalledWith({
        type: 'click',
        id: 'menu_item_file_open',
      })
    })
  })

  // A key held down repeats, and a menu item is picked once per press. An IME
  // reports the whole composition as one keydown of its own, which is a
  // composition rather than a keystroke.
  test('a repeat and a composition are not presses', () => {
    const { update } = renderApp({ menu: MENU })

    press(document.body,
      { key: 'o', code: 'KeyO', ctrlKey: true, repeat: true })
    press(document.body,
      { key: 'o', code: 'KeyO', ctrlKey: true, isComposing: true })

    expect(update).not.toHaveBeenCalled()
  })

  // AltGr is Control and Alt held together on Windows and on some layouts
  // elsewhere, so the third character on a key arrives looking exactly like a
  // Ctrl+Alt chord. The character wins: firing the item would eat a keystroke
  // the visitor meant to type.
  test('AltGr is a character, not a chord', () => {
    const menu = [{
      type: 'text', label: 'Euro', id: 'menu_item_euro',
      accelerator: 'Ctrl+OptionOrAlt+e',
    }]

    const { update } = renderApp({ menu })

    // Without the marker the same event is the chord it looks like.
    press(document.body,
      { key: 'e', code: 'KeyE', ctrlKey: true, altKey: true })
    expect(update).toHaveBeenCalledWith({ type: 'click', id: 'menu_item_euro' })
    update.mockClear()

    expect(press(document.body, {
      key: '€', code: 'KeyE', ctrlKey: true, altKey: true,
      modifierAltGraph: true,
    })).toBe(true)
    expect(update).not.toHaveBeenCalled()
  })

  // The whole point of the listener being conditional: a menu that declares
  // no accelerator does not listen for a keystroke at all.
  describe('a menu with no accelerator', () => {
    // The same tree with the declarations taken off, so the two differ in
    // nothing else.
    const BARE = MENU.map((node) => node.type === 'submenu'
      ? { ...node, children: node.children.map(({ accelerator, ...n }) => n) }
      : (({ accelerator, ...n }) => n)(node))

    // How many keydown listeners the document collects while menu is drawn.
    // A count rather than a check for none: Mantine's own dropdowns bind one
    // of their own, so what says whether the menubar bound anything is the
    // difference between the two trees.
    function keydownListeners(menu) {
      const add = vi.spyOn(document, 'addEventListener')
      renderApp({ menu })

      const bound = add.mock.calls.filter(([type]) => type === 'keydown')
      add.mockRestore()
      cleanup()

      return bound.length
    }

    test('binds nothing', () => {
      expect(keydownListeners(BARE)).toBe(keydownListeners(MENU) - 1)
    })

    test('lets every keystroke through', () => {
      const { update } = renderApp({ menu: BARE })

      // Not cancelled either: the combination stays the browser's.
      expect(press(document.body, { key: 'o', code: 'KeyO', ctrlKey: true }))
        .toBe(true)
      expect(update).not.toHaveBeenCalled()
    })
  })

  test('the listener goes with the menubar', () => {
    const remove = vi.spyOn(document, 'removeEventListener')

    const { unmount } = renderApp({ menu: MENU })
    unmount()

    expect(remove.mock.calls.filter(([type]) => type === 'keydown').length)
      .toBeGreaterThan(0)
  })

  // The item says how else to reach it, in the spelling the platform uses.
  test('an item shows its accelerator', async () => {
    const { unmount } = renderApp({ menu: MENU })

    screen.getByRole('button', { name: /^File/ }).click()
    await screen.findByText('Open')
    expect(screen.getByText('Ctrl+O')).toBeInTheDocument()
    expect(screen.getByText('Ctrl+Shift+S')).toBeInTheDocument()

    unmount()
    asMac(true)
    renderApp({ menu: MENU })

    screen.getByRole('button', { name: /^File/ }).click()
    await screen.findByText('Open')
    expect(screen.getByText('⌘O')).toBeInTheDocument()
    expect(screen.getByText('⇧⌘S')).toBeInTheDocument()
  })
})
