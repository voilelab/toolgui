import React from 'react'
import { act, cleanup, render, screen, waitFor } from '@testing-library/react'
import { afterEach, expect, test, describe, vi } from 'vitest'

import { SessionApp } from '@toolgui-web/lib'

function appConf(hashMode) {
  return {
    page_names: ['problems', 'detail'],
    page_confs: {
      problems: { name: 'problems', title: 'Problems', emoji: '' },
      detail: { name: 'detail', title: 'Detail', emoji: '' },
    },
    title: '',
    main_container_id: 'container_main',
    sidebar_container_id: 'container_sidebar',
    hash_page_name_mode: hashMode,
    version: 'v0.0.0',
    show_version: false,
  }
}

// fakeBackend records what SessionApp asks of a backend.
function fakeBackend(conf) {
  let onPack
  const backend = {
    appConf: vi.fn(async () => conf),
    start: vi.fn(async () => {}),
    update: vi.fn(async () => {}),
    uploadFile: vi.fn(),
    downloadFile: vi.fn(),
  }

  return {
    backend,
    connect: (cb) => { onPack = cb; return backend },
    pack: (p) => act(() => { onPack(p) }),
  }
}

afterEach(() => {
  cleanup()
  window.location.hash = ''
})

describe('SessionApp', () => {
  test('opens the first page, and a navigate pack opens a new session', async () => {
    const fake = fakeBackend(appConf(false))
    render(<SessionApp connect={fake.connect} />)

    await waitFor(() => expect(fake.backend.start).toHaveBeenCalledWith('problems', ''))

    fake.pack({ navigate: { page: 'detail', query: { id: ['0004'] } } })
    await waitFor(() => expect(fake.backend.start).toHaveBeenCalledWith('detail', 'id=0004'))
    expect(window.location.hash).toBe('')
  })

  test('useHash reads the page off the hash and follows it', async () => {
    window.location.hash = '#/detail?id=1'

    const fake = fakeBackend(appConf(true))
    render(<SessionApp connect={fake.connect} useHash />)

    await waitFor(() => expect(fake.backend.start).toHaveBeenCalledWith('detail', 'id=1'))

    // In hash mode a navigate moves the hash, and the hash opens the page.
    fake.pack({ navigate: { page: 'problems', query: {} } })
    await waitFor(() => expect(fake.backend.start).toHaveBeenCalledWith('problems', ''))
    expect(window.location.hash).toBe('#/problems')
  })

  test('shows loading until the app config, and a failed start', async () => {
    const fake = fakeBackend(appConf(false))
    let release
    fake.backend.appConf.mockImplementation(() => new Promise((r) => { release = r }))
    fake.backend.start.mockRejectedValue(new Error('page not found'))
    const error = vi.spyOn(console, 'error').mockImplementation(() => {})

    render(<SessionApp connect={fake.connect} loading={<p>loading…</p>} />)
    expect(screen.getByText('loading…')).toBeTruthy()

    await act(async () => { release(appConf(false)) })
    await waitFor(() => expect(screen.getByText(/page not found/)).toBeTruthy())
    error.mockRestore()
  })
})
