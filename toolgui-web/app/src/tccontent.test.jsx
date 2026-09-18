import React from 'react'
import { cleanup } from '@testing-library/react'

import { render } from './render'
import { afterEach, expect, test, describe, vi } from 'vitest'

import { Node } from '@toolgui-web/lib/src/app/Nodes'
import { TCode } from '@toolgui-web/lib/src/components/tccontent/code'
import { TMarkdown } from '@toolgui-web/lib/src/components/tccontent/markdown'
import { TTable } from '@toolgui-web/lib/src/components/tcdata/table'

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

describe('markdown gfm', () => {
  test('a table is the table a Table component draws', () => {
    const own = render(
      <TTable node={new Node('main/0', {
        name: 'table_component', id: '',
        head: ['a', 'b'], table: [['1', '2']],
      })} {...RENDER_PROPS} />
    ).container.querySelector('table')
    const { container } = markdown('| a | b |\n| - | - |\n| 1 | 2 |')
    const table = container.querySelector('table')

    expect(table).not.toBeNull()
    // Mantine's own classes are what carry the table's look; the markdown one
    // adds a class of its own to scope the rules it resets.
    for (const c of own.classList) expect(table.classList).toContain(c)
    expect([...table.querySelectorAll('th')].map(e => e.textContent))
      .toEqual(['a', 'b'])
    expect([...table.querySelectorAll('td')].map(e => e.textContent))
      .toEqual(['1', '2'])
  })

  test('a column keeps its alignment', () => {
    const { container } = markdown(
      '| l | c | r |\n| :- | :-: | -: |\n| 1 | 2 | 3 |')
    expect([...container.querySelectorAll('th')].map(e => e.style.textAlign))
      .toEqual(['left', 'center', 'right'])
  })

  test('a shortcode in a cell expands', () => {
    const { container } = markdown('| a |\n| - |\n| pear :tada: |')
    expect(container.querySelector('td').textContent).toBe('pear 🎉')
  })

  test('strikethrough and a task list render', () => {
    const { container } = markdown('~~gone~~\n\n- [x] done\n- [ ] todo')
    expect(container.querySelector('del').textContent).toBe('gone')
    const boxes = [...container.querySelectorAll('input[type=checkbox]')]
    expect(boxes.map(e => e.checked)).toEqual([true, false])
  })
})
