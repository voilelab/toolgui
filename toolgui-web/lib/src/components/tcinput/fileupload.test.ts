import { expect, test, vi } from 'vitest'

import { clearState } from '../state'
import { acceptLabel, formatSize, uploadPick } from './fileupload'
import { UploadResult } from '../../app/Upload'

test('clearState mid-pick stops further uploads and sends', async () => {
  const files = [new File(['a'], 'a'), new File(['b'], 'b'), new File(['c'], 'c')]

  let finish: (r: UploadResult) => void = () => { }
  const upload = vi.fn(() => new Promise<UploadResult>(r => { finish = r }))
  const send = vi.fn()

  const done = uploadPick('f', files, upload, send, () => { })
  await vi.waitFor(() => expect(upload).toHaveBeenCalledTimes(1))

  clearState()
  finish({ ok: true })
  await done

  expect(upload).toHaveBeenCalledTimes(1)
  // Only the clear sent before the session changed.
  expect(send.mock.calls).toEqual([[[]]])
})

test('formatSize', () => {
  expect(formatSize(0)).toBe('0 B')
  expect(formatSize(1023)).toBe('1023 B')
  expect(formatSize(1536)).toBe('1.5 KB')
  expect(formatSize(200 * 1024 * 1024)).toBe('200.0 MB')
})

test('acceptLabel', () => {
  expect(acceptLabel(undefined)).toBe('')
  expect(acceptLabel('.csv')).toBe('CSV')
  expect(acceptLabel('.csv, image/png,')).toBe('CSV, image/png')
})
