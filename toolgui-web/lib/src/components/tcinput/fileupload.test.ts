import { expect, test, vi } from 'vitest'

import { clearState } from '../state'
import { uploadPick } from './fileupload'
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
