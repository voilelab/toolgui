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

    await waitFor(() => expect(update).toHaveBeenCalled())

    expect(upload.mock.calls.map((c) => c[1])).toEqual(['up/0', 'up/1'])
    expect(update).toHaveBeenCalledWith({
      type: 'input',
      id: 'up',
      value: [
        { name: 'a.txt', type: 'text/plain', size: 1 },
        { name: 'b.txt', type: 'text/plain', size: 2 },
      ],
    })
  })
})
