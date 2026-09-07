import React from 'react'
import { act, render } from '@testing-library/react'
import { expect, test, vi } from 'vitest'

import { TPlugin } from '@toolgui-web/lib/src/components/tcmisc/plugin'

const PROPS = {
  id: 'plugin_component_my_plugin',
  src: '/plugin/gauge/gauge.js',
  style: '',
  props: { value: 42 },
  width: '100%',
  height: '150px',
}

// mountPlugin renders the component and hands back a sender that posts as the
// plugin would, with event.source set to the frame's own window.
function mountPlugin(overrides = {}) {
  const update = vi.fn()
  const upload = vi.fn()

  const { container } = render(
    <TPlugin
      node={{ props: { ...PROPS, ...overrides } }}
      update={update}
      upload={upload}
      theme="dark" />
  )

  const iframe = container.querySelector('iframe')

  const sendAsGuest = (data, source) => act(() => {
    window.dispatchEvent(new MessageEvent('message', {
      data: data,
      source: source === undefined ? iframe.contentWindow : source,
    }))
  })

  return { iframe, update, upload, sendAsGuest }
}

test('the plugin runs on an opaque origin, like any other guest', () => {
  expect(mountPlugin().iframe.getAttribute('sandbox')).toBe('allow-scripts')
})

test('the guest document is the bridge plus a classic script tag', () => {
  const srcdoc = mountPlugin().iframe.getAttribute('srcdoc')

  expect(srcdoc).toContain('window.toolgui')
  // Deferred, so the plugin runs with a document that has a body.
  expect(srcdoc).toContain('<script defer src="/plugin/gauge/gauge.js"></script>')

  // A module is fetched with cors, which an opaque-origin frame cannot pass.
  expect(srcdoc).not.toContain('type="module"')

  // The bridge has to be installed before the plugin runs.
  expect(srcdoc.indexOf('window.toolgui'))
    .toBeLessThan(srcdoc.indexOf('gauge.js'))
})

test('a stylesheet is loaded only when the plugin asks for one', () => {
  expect(mountPlugin().iframe.getAttribute('srcdoc')).not.toContain('<link')

  expect(mountPlugin({ style: '/plugin/gauge/gauge.css' }).iframe.getAttribute('srcdoc'))
    .toContain('<link rel="stylesheet" href="/plugin/gauge/gauge.css">')
})

test('a url cannot close the attribute it sits in', () => {
  const srcdoc = mountPlugin({ src: '/plugin/x.js"><script>bad()</script>' })
    .iframe.getAttribute('srcdoc')

  expect(srcdoc).not.toContain('bad()</script>')
  expect(srcdoc).toContain('&quot;')
})

test('ready is answered with a render carrying the props from go', () => {
  const { iframe, sendAsGuest } = mountPlugin()

  const post = vi.spyOn(iframe.contentWindow, 'postMessage')
  sendAsGuest({ toolgui: 1, type: 'ready' })

  const [message] = post.mock.calls[0]

  expect(message.type).toBe('render')
  expect(message.id).toBe(PROPS.id)
  expect(message.theme).toBe('dark')
  expect(message.props).toEqual({ value: 42 })
})

test('an update is stamped with the plugin own id, not one the guest chose', () => {
  const { update, sendAsGuest } = mountPlugin()

  sendAsGuest({ toolgui: 1, type: 'update', id: 'button_component_admin', value: 7 })

  expect(update).toHaveBeenCalledWith({
    type: 'custom',
    id: PROPS.id,
    value: 7,
  })
})

test('a plugin with no id has no state to write to, so it writes nothing', () => {
  const { update, sendAsGuest } = mountPlugin({ id: '' })

  sendAsGuest({ toolgui: 1, type: 'update', value: 7 })

  expect(update).not.toHaveBeenCalled()
})
