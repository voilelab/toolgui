import React from 'react'
import { cleanup, screen } from '@testing-library/react'
import { afterEach, expect, test, describe, vi } from 'vitest'

import { render } from './render'

import { Node } from '@toolgui-web/lib/src/app/Nodes'
import { TSpinner } from '@toolgui-web/lib/src/components/tcmisc/spinner'

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
