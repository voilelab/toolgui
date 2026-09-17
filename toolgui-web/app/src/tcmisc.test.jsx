import React from 'react'
import { cleanup, screen } from '@testing-library/react'
import { afterEach, expect, test, describe, vi } from 'vitest'
import { Notifications, cleanNotifications } from '@mantine/notifications'

import { render } from './render'

import { Node } from '@toolgui-web/lib/src/app/Nodes'
import { TSpinner } from '@toolgui-web/lib/src/components/tcmisc/spinner'
import { TToast } from '@toolgui-web/lib/src/components/tcmisc/toast'

const RENDER_PROPS = { update: vi.fn(), upload: vi.fn(), theme: 'light' }

// Vitest runs without globals, so RTL's auto-cleanup never registers.
afterEach(cleanup)

describe('TSpinner', () => {
  test('shows the label next to the spinner', () => {
    const { container } = render(
      <TSpinner
        node={new Node('main/0', {
          name: 'spinner_component', id: '', label: 'Working…',
        })}
        {...RENDER_PROPS} />
    )

    expect(screen.getByText('Working…')).toBeVisible()
    expect(container.querySelector('.mantine-Loader-root')).not.toBeNull()
  })
})

// toastNode is the node a Toast on run `seq` sends, at one fixed position.
function toastNode(seq, props) {
  return new Node('main/0', {
    name: 'toast_component', id: '',
    text: 'Saved', icon: '', duration_ms: 0, seq,
    ...props,
  })
}

// page is what the app renders around a toast: the notifications container,
// and the place on the page where the page function wrote the Toast call.
function page(node) {
  return (
    <>
      <Notifications />
      <div data-testid="slot">
        <TToast node={node} {...RENDER_PROPS} />
      </div>
    </>
  )
}

describe('TToast', () => {
  // The notifications store is module state, so a toast left over from one
  // test would be counted by the next. Wrapped rather than passed straight to
  // afterEach, which would hand it the test context as the store to clean.
  afterEach(() => cleanNotifications())

  test('takes up no room on the page', () => {
    render(page(toastNode(1)))

    // The toast is drawn by the notifications container, over the page, so
    // the place the page function wrote it holds nothing.
    expect(screen.getByTestId('slot').childNodes).toHaveLength(0)
    expect(screen.getByText('Saved')).toBeVisible()
  })

  test('shows the icon in front of the text', () => {
    render(page(toastNode(1, { icon: '✅' })))

    expect(screen.getByText('✅')).toBeVisible()
  })

  // The one the ticket is about. Two runs write the same Toast call in the
  // same place, so the node and its React key are the same both times and
  // nothing remounts. The run serial is what has to fire it again.
  test('fires again when the next run writes the same toast', async () => {
    const { rerender } = render(page(toastNode(1)))
    expect(screen.getAllByText('Saved')).toHaveLength(1)

    rerender(page(toastNode(2)))
    expect(await screen.findAllByText('Saved')).toHaveLength(2)
  })

  // An interrupted run leaves its toast node in the tree until a later run
  // drops it. The toast it already fired is the notifications container's now,
  // so dropping the node neither takes that toast back nor leaves anything on
  // the page.
  test('dropping the node leaves nothing behind', () => {
    const { rerender } = render(page(toastNode(1)))
    expect(screen.getByText('Saved')).toBeVisible()

    rerender(<Notifications />)

    expect(screen.queryByTestId('slot')).toBeNull()
    expect(screen.getByText('Saved')).toBeVisible()
  })

  // A rerender that is not a new run -- the page around the toast changed,
  // the theme flipped -- must not fire it a second time.
  test('does not fire again on a rerender of the same run', () => {
    const { rerender } = render(page(toastNode(1)))

    rerender(page(toastNode(1)))
    expect(screen.getAllByText('Saved')).toHaveLength(1)
  })
})
