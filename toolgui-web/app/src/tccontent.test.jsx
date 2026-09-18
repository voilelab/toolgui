import React from 'react'
import { cleanup } from '@testing-library/react'

import { render } from './render'
import { afterEach, expect, test, describe, vi } from 'vitest'

import { Node } from '@toolgui-web/lib/src/app/Nodes'
import { TCode } from '@toolgui-web/lib/src/components/tccontent/code'
import { TMarkdown } from '@toolgui-web/lib/src/components/tccontent/markdown'

const RENDER_PROPS = { update: vi.fn(), upload: vi.fn(), theme: 'light' }

// Vitest runs without globals, so RTL's auto-cleanup never registers.
afterEach(cleanup)

const code = (props) => render(
  <TCode node={new Node('main/0', { name: 'code_component', id: '', ...props })}
    {...RENDER_PROPS} />
)

const markdown = (text) => render(
  <TMarkdown node={new Node('main/0', { name: 'markdown_component', id: '', text })}
    {...RENDER_PROPS} />
)

describe('code blocks', () => {
  test('a Markdown fence is the block a Code component draws', () => {
    const own = code({ code: 'func main() {}', lang: 'go' })
      .container.querySelector('pre.toolgui-code')
    const fenced = markdown('```go\nfunc main() {}\n```')
      .container.querySelector('pre.toolgui-code')

    expect(own).not.toBeNull()
    expect(fenced).not.toBeNull()
    expect(fenced.getAttribute('style')).toBe(own.getAttribute('style'))
    expect(fenced.querySelector('code').getAttribute('style'))
      .toBe(own.querySelector('code').getAttribute('style'))
    expect(fenced.textContent).toBe(own.textContent)
  })

  test('a fence is not wrapped in a second pre', () => {
    const { container } = markdown('```go\nfunc main() {}\n```')
    expect(container.querySelectorAll('pre')).toHaveLength(1)
  })

  test('a fence without a language still gets the block', () => {
    const { container } = markdown('```\nplain\n```')
    const pre = container.querySelector('pre.toolgui-code')
    expect(pre).not.toBeNull()
    expect(pre.textContent).toBe('plain')
  })

  test('an inline code span stays inline', () => {
    const { container } = markdown('text with `inline` in it')
    expect(container.querySelector('pre')).toBeNull()
    expect(container.querySelector('code').textContent).toBe('inline')
  })
})
