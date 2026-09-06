import React from 'react'
import { act, render } from '@testing-library/react'
import { expect, test, vi } from 'vitest'

import { TIframe } from '@toolgui-web/lib/src/components/tcmisc/iframe'

const PROPS = {
  id: 'iframe_component_my_iframe',
  html: '<b>hi</b>',
  script: true,
  width: '100%',
  height: '150px',
}

// mountIframe renders the component and hands back a sender that posts as the
// guest would, with event.source set to the frame's own window.
function mountIframe(overrides = {}) {
  const update = vi.fn()
  const upload = vi.fn()

  const { container } = render(
    <TIframe
      node={{ props: { ...PROPS, ...overrides } }}
      update={update}
      upload={upload}
      theme="dark" />
  )

  const iframe = container.querySelector('iframe')

  // act() so a message that drives state (resize) has re-rendered by the
  // time the assertion runs.
  const sendAsGuest = (data, source) => act(() => {
    window.dispatchEvent(new MessageEvent('message', {
      data: data,
      source: source === undefined ? iframe.contentWindow : source,
    }))
  })

  return { iframe, update, upload, sendAsGuest }
}

test('the guest helper is prepended to srcdoc so it needs no cross-origin access', () => {
  const { iframe } = mountIframe()

  expect(iframe.getAttribute('srcdoc')).toContain('window.toolgui')
  expect(iframe.getAttribute('srcdoc')).toContain('<b>hi</b>')
})

test('no helper is shipped when the guest cannot run scripts', () => {
  const { iframe } = mountIframe({ script: false })

  expect(iframe.getAttribute('srcdoc')).toBe('<b>hi</b>')
})

test('ready is answered with a render carrying the props and theme', () => {
  const { iframe, sendAsGuest } = mountIframe()

  const post = vi.spyOn(iframe.contentWindow, 'postMessage')
  sendAsGuest({ toolgui: 1, type: 'ready' })

  expect(post).toHaveBeenCalledTimes(1)
  const [message, targetOrigin] = post.mock.calls[0]

  expect(message.type).toBe('render')
  expect(message.toolgui).toBe(1)
  expect(message.id).toBe(PROPS.id)
  expect(message.theme).toBe('dark')
  // The guest is that html; sending it back would be a second copy.
  expect(message.props.html).toBeUndefined()
  // An opaque-origin guest has no origin to name.
  expect(targetOrigin).toBe('*')
})

test('an update is stamped with the iframe own id, not one the guest chose', () => {
  const { update, sendAsGuest } = mountIframe()

  sendAsGuest({ toolgui: 1, type: 'update', id: 'button_component_admin', value: { clicked: true } })

  expect(update).toHaveBeenCalledWith({
    type: 'iframe',
    id: PROPS.id,
    value: { clicked: true },
  })
})

test('a message from another window is ignored', () => {
  const { update, sendAsGuest } = mountIframe()

  sendAsGuest({ toolgui: 1, type: 'update', value: { clicked: true } }, window)

  expect(update).not.toHaveBeenCalled()
})

test('a message without the protocol marker is ignored', () => {
  const { update, sendAsGuest } = mountIframe()

  sendAsGuest({ type: 'update', value: { clicked: true } })

  expect(update).not.toHaveBeenCalled()
})

test('resize drives the height only when the component asked for auto', () => {
  const auto = mountIframe({ height: 'auto' })
  auto.sendAsGuest({ toolgui: 1, type: 'resize', height: 320 })
  expect(auto.iframe.style.height).toBe('320px')

  const fixed = mountIframe({ height: '150px' })
  fixed.sendAsGuest({ toolgui: 1, type: 'resize', height: 320 })
  expect(fixed.iframe.style.height).toBe('150px')
})
