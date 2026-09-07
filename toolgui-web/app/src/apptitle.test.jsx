import React from 'react'
import { cleanup, render } from '@testing-library/react'
import { afterEach, beforeEach, expect, test, describe, vi } from 'vitest'

import { App } from '@toolgui-web/lib'

function appConf(title) {
  return {
    page_names: ['index'],
    page_confs: { index: { name: 'index', title: 'Index', emoji: '' } },
    title: title,
    main_container_id: 'container_main',
    sidebar_container_id: 'container_sidebar',
    hash_page_name_mode: true,
    version: 'v0.0.0',
    show_version: false,
  }
}

const RENDER_PROPS = { update: vi.fn(), upload: vi.fn() }

// index.html carries the icon link the app rewrites, so the test document
// needs one too.
beforeEach(() => {
  const link = document.createElement('link')
  link.setAttribute('rel', 'icon')
  document.head.appendChild(link)
})

afterEach(() => {
  cleanup()
  document.head.querySelectorAll("link[rel='icon']").forEach(l => l.remove())
})

describe('document title', () => {
  test('holds the page title and the app title', () => {
    render(<App appConf={appConf('My Tool')} pageName="index" {...RENDER_PROPS} />)
    expect(document.title).toBe('Index - My Tool')
  })

  test('is the page title alone when the app has no title', () => {
    render(<App appConf={appConf('')} pageName="index" {...RENDER_PROPS} />)
    expect(document.title).toBe('Index')
  })

  test('names the app on a page it doesn\'t have', () => {
    render(<App appConf={appConf('My Tool')} pageName="nope" {...RENDER_PROPS} />)
    expect(document.title).toBe('Page not found - My Tool')
  })
})
