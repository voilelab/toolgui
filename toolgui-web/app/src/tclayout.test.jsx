import React from 'react'
import { cleanup, render, screen, fireEvent } from '@testing-library/react'
import { afterEach, expect, test, describe, vi } from 'vitest'

import { Node } from '@toolgui-web/lib/src/app/Nodes'
import { TTab } from '@toolgui-web/lib/src/components/tclayout/tab'
import { TExpand } from '@toolgui-web/lib/src/components/tclayout/expand'

function node(key, props, children = []) {
  const n = new Node(key, props)
  n.children = children
  return n
}

const text = (key, t) => node(key, { name: 'text_component', id: '', text: t })

const RENDER_PROPS = { update: vi.fn(), upload: vi.fn(), theme: 'light' }

// Vitest runs without globals, so RTL's auto-cleanup never registers.
afterEach(cleanup)

describe('TTab', () => {
  const mountTab = () => render(
    <TTab
      node={node('main/0', { name: 'tab_component', id: '', tabs: ['tab1', 'tab2'] }, [
        text('main/0/0', 'A tab!'),
        text('main/0/1', 'B tab!'),
      ])}
      {...RENDER_PROPS} />
  )

  test('shows the active tab and hides the rest', () => {
    mountTab()

    expect(screen.getByText('A tab!')).toBeVisible()
    expect(screen.getByText('B tab!')).not.toBeVisible()

    fireEvent.click(screen.getByText('tab2'))

    expect(screen.getByText('A tab!')).not.toBeVisible()
    expect(screen.getByText('B tab!')).toBeVisible()
  })

  test('keeps the tab you leave mounted', () => {
    mountTab()

    const content = screen.getByText('A tab!')

    fireEvent.click(screen.getByText('tab2'))
    fireEvent.click(screen.getByText('tab1'))

    // The same DOM node, so nothing a tab holds was rebuilt.
    expect(screen.getByText('A tab!')).toBe(content)
  })
})

describe('TExpand', () => {
  const mountExpand = (expanded) => render(
    <TExpand
      node={node('main/0', { name: 'expand_component', id: '', title: 'Expand', expanded }, [
        text('main/0/0', 'A expand!'),
      ])}
      {...RENDER_PROPS} />
  )

  test('does not build the contents until the first open', () => {
    mountExpand(false)

    expect(screen.queryByText('A expand!')).toBeNull()

    fireEvent.click(screen.getByText('Expand'))

    expect(screen.getByText('A expand!')).toBeVisible()
  })

  test('opens from the keyboard', () => {
    mountExpand(false)

    const header = screen.getByRole('button', { name: /Expand/ })
    expect(header).toHaveAttribute('aria-expanded', 'false')

    fireEvent.keyDown(header, { key: 'Enter' })

    expect(screen.getByText('A expand!')).toBeVisible()
    expect(header).toHaveAttribute('aria-expanded', 'true')
  })

  test('keeps the contents mounted once opened', () => {
    mountExpand(true)

    const content = screen.getByText('A expand!')

    fireEvent.click(screen.getByText('Expand'))
    expect(screen.getByText('A expand!')).not.toBeVisible()

    fireEvent.click(screen.getByText('Expand'))
    expect(screen.getByText('A expand!')).toBe(content)
  })
})
