import React from 'react'
import { act, render, screen } from '@testing-library/react'
import { afterEach, beforeEach, expect, test, vi } from 'vitest'

import { WSApp } from './wsapp'

const APP_CONF = {
  page_names: ['index'],
  page_confs: { index: { name: 'index', title: 'Index' } },
  hash_page_name_mode: false,
  version: 'v0.0.0',
  show_version: true,
  main_container_id: 'container_main',
  sidebar_container_id: 'container_sidebar',
}

let sockets

beforeEach(() => {
  sockets = []

  // index.html ships one, and App sets the page emoji on it.
  const icon = document.createElement('link')
  icon.setAttribute('rel', 'icon')
  document.head.appendChild(icon)

  // Record every attempt, and never open: what matters here is how many the
  // app makes.
  vi.stubGlobal('WebSocket', class {
    constructor(url) {
      this.url = url
      sockets.push(this)
    }
    send() { }
    close() { }
  })

  vi.stubGlobal('fetch', vi.fn(async () => ({
    status: 200,
    ok: true,
    json: async () => APP_CONF,
  })))
})

afterEach(() => {
  document.querySelector(`head > link[rel='icon']`).remove()
  vi.unstubAllGlobals()
})

async function renderAt(path) {
  window.history.pushState({}, '', path)
  await act(async () => { render(<WSApp />) })
}

test('a known page connects', async () => {
  await renderAt('/index')

  expect(sockets.map(s => s.url)).toEqual(['ws://localhost:3000/api/update/index'])
})

test('an unknown page renders not found without connecting', async () => {
  await renderAt('/main')

  expect(sockets).toHaveLength(0)
  expect(screen.getByText(/Page not found/i)).toBeInTheDocument()
})
