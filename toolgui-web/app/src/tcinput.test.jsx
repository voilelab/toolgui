import React from 'react'
import { cleanup, fireEvent, waitFor } from '@testing-library/react'
import { afterEach, expect, test, describe, vi } from 'vitest'

import { render } from './render'

import { Node } from '@toolgui-web/lib/src/app/Nodes'
import { TFileupload } from '@toolgui-web/lib/src/components/tcinput/fileupload'

// Vitest runs without globals, so RTL's auto-cleanup never registers.
afterEach(cleanup)

describe('TFileupload', () => {
  test('uploads each picked file under its own key', async () => {
    const update = vi.fn()
    const upload = vi.fn().mockResolvedValue({ ok: true })

    const { container } = render(
      <TFileupload
        node={new Node('main/0', {
          name: 'fileupload_component', id: 'up', label: 'Files',
          accept: '', disabled: false, multiple: true,
        })}
        update={update} upload={upload} theme="light" />
    )

    const a = new File(['a'], 'a.txt', { type: 'text/plain' })
    const b = new File(['bb'], 'b.txt', { type: 'text/plain' })
    fireEvent.change(container.querySelector('input[type=file]'),
      { target: { files: [a, b] } })

    await waitFor(() => expect(update).toHaveBeenCalledTimes(2))

    expect(upload.mock.calls.map((c) => c[1])).toEqual(['up/0', 'up/1'])
    // The old pick is cleared before any file of the new one is stored.
    expect(update).toHaveBeenNthCalledWith(1,
      { type: 'input', id: 'up', value: [] })
    expect(update).toHaveBeenLastCalledWith({
      type: 'input',
      id: 'up',
      value: [
        { name: 'a.txt', type: 'text/plain', size: 1 },
        { name: 'b.txt', type: 'text/plain', size: 2 },
      ],
    })
  })

  test('refuses a pick over the limit', async () => {
    const update = vi.fn()
    const upload = vi.fn().mockResolvedValue({ ok: true })

    const { container } = render(
      <TFileupload
        node={new Node('main/0', {
          name: 'fileupload_component', id: 'many', label: 'Files',
          accept: '', disabled: false, multiple: true,
        })}
        update={update} upload={upload} theme="light" />
    )

    const files = Array.from({ length: 1001 },
      (_, i) => new File(['x'], `${i}.txt`))
    fireEvent.change(container.querySelector('input[type=file]'),
      { target: { files } })

    await waitFor(() =>
      expect(container.textContent).toContain('Pick at most 1000 files'))
    expect(upload).not.toHaveBeenCalled()
    expect(update).not.toHaveBeenCalled()
  })

  test('stops an old pick once a new one starts', async () => {
    const update = vi.fn()
    let release
    const first = new Promise((r) => { release = r })
    const upload = vi.fn()
      .mockReturnValueOnce(first)
      .mockResolvedValue({ ok: true })

    const { container } = render(
      <TFileupload
        node={new Node('main/0', {
          name: 'fileupload_component', id: 'race', label: 'Files',
          accept: '', disabled: false, multiple: true,
        })}
        update={update} upload={upload} theme="light" />
    )

    const input = container.querySelector('input[type=file]')
    fireEvent.change(input, { target: { files: [
      new File(['a'], 'a.txt'), new File(['b'], 'b.txt')] } })
    await waitFor(() => expect(upload).toHaveBeenCalledTimes(1))

    fireEvent.change(input, { target: { files: [new File(['c'], 'c.txt')] } })
    release({ ok: true })

    await waitFor(() => expect(update).toHaveBeenCalledTimes(3))

    // The old pick never uploads b.txt, and never sends its value.
    expect(upload.mock.calls.map((c) => [c[0].name, c[1]])).toEqual([
      ['a.txt', 'race/0'], ['c.txt', 'race/0']])
    expect(update).toHaveBeenLastCalledWith({
      type: 'input', id: 'race',
      value: [{ name: 'c.txt', type: '', size: 1 }],
    })
  })
})
