import React from 'react'
import { cleanup, screen } from '@testing-library/react'

import { render } from './render'
import { afterEach, expect, test, describe, vi } from 'vitest'

import { Node } from '@toolgui-web/lib/src/app/Nodes'
import { emojize } from '@toolgui-web/lib/src/util/emoji'
import { TText } from '@toolgui-web/lib/src/components/tccontent/text'
import { TTitle } from '@toolgui-web/lib/src/components/tccontent/title'
import { TSubtitle } from '@toolgui-web/lib/src/components/tccontent/subtitle'
import { TLink } from '@toolgui-web/lib/src/components/tccontent/link'
import { TCode } from '@toolgui-web/lib/src/components/tccontent/code'
import { TMarkdown } from '@toolgui-web/lib/src/components/tccontent/markdown'

function node(key, props) {
  return new Node(key, props)
}

const RENDER_PROPS = { update: vi.fn(), upload: vi.fn(), theme: 'light' }

// Vitest runs without globals, so RTL's auto-cleanup never registers.
afterEach(cleanup)

describe('emojize', () => {
  test('replaces a known shortcode', () => {
    expect(emojize('ship it :tada:')).toBe('ship it 🎉')
  })

  test('replaces every shortcode in the text', () => {
    expect(emojize(':+1: and :-1:')).toBe('👍 and 👎')
  })

  test('replaces adjacent shortcodes', () => {
    expect(emojize(':cat::dog:')).toBe('🐱🐶')
  })

  test('keeps an unknown shortcode as written', () => {
    expect(emojize('a :notanemoji: b')).toBe('a :notanemoji: b')
  })

  test('keeps a name inherited from Object', () => {
    expect(emojize(':constructor:')).toBe(':constructor:')
    expect(emojize(':toString:')).toBe(':toString:')
    expect(emojize(':__proto__:')).toBe(':__proto__:')
  })

  test('expands the all-digit shortcodes', () => {
    expect(emojize('this :100:')).toBe('this 💯')
    expect(emojize(':1234:')).toBe('🔢')
  })

  test('leaves a clock time alone', () => {
    expect(emojize('runs at 10:30:00')).toBe('runs at 10:30:00')
    // The demo renders this one, so the e2e run depends on it.
    expect(emojize('2006-01-02 15:04:05')).toBe('2006-01-02 15:04:05')
  })

  test('leaves text with no colon alone', () => {
    expect(emojize('nothing to do here')).toBe('nothing to do here')
  })

  test('does not carry regex state between calls', () => {
    expect(emojize(':tada: :tada:')).toBe('🎉 🎉')
    expect(emojize(':tada:')).toBe('🎉')
  })
})

describe('content components', () => {
  test('TText expands', () => {
    render(<TText node={node('main/0', { name: 'text_component', id: '', text: 'hi :wave:' })}
      {...RENDER_PROPS} />)
    expect(screen.getByText('hi 👋')).toBeInTheDocument()
  })

  test('TTitle expands', () => {
    render(<TTitle node={node('main/0', { name: 'title_component', id: '', text: ':rocket: Launch' })}
      {...RENDER_PROPS} />)
    expect(screen.getByText('🚀 Launch')).toBeInTheDocument()
  })

  test('TSubtitle expands', () => {
    render(<TSubtitle node={node('main/0', { name: 'subtitle_component', id: '', text: ':books: Docs' })}
      {...RENDER_PROPS} />)
    expect(screen.getByText('📚 Docs')).toBeInTheDocument()
  })

  test('TLink expands its text but not its url', () => {
    render(<TLink node={node('main/0', {
      name: 'link_component', id: '', text: ':star: toolgui',
      url: 'https://example.com/:star:',
    })} {...RENDER_PROPS} />)
    const link = screen.getByRole('link', { name: '⭐ toolgui' })
    expect(link).toHaveAttribute('href', 'https://example.com/:star:')
  })

  test('TCode keeps a shortcode literal', () => {
    render(<TCode node={node('main/0', {
      name: 'code_component', id: '', code: 'print(":tada:")', lang: 'python',
    })} {...RENDER_PROPS} />)
    expect(screen.getByText(/:tada:/)).toBeInTheDocument()
  })
})

describe('TMarkdown', () => {
  const markdown = (text) => render(
    <TMarkdown node={node('main/0', { name: 'markdown_component', id: '', text })}
      {...RENDER_PROPS} />
  )

  test('expands in prose', () => {
    markdown('ship it :tada:')
    expect(screen.getByText('ship it 🎉')).toBeInTheDocument()
  })

  test('expands inside a heading and a list item', () => {
    markdown('# :rocket: Launch\n\n* first :100:')
    expect(screen.getByRole('heading', { name: '🚀 Launch' })).toBeInTheDocument()
    expect(screen.getByRole('listitem')).toHaveTextContent('first 💯')
  })

  test('keeps a shortcode in a code span literal', () => {
    const { container } = markdown('write `:tada:` for it')
    expect(container.querySelector('code').textContent).toBe(':tada:')
  })

  test('keeps a shortcode in a fenced block literal', () => {
    const { container } = markdown('```\nprint(":tada:")\n```')
    expect(container.textContent).toContain(':tada:')
    expect(container.textContent).not.toContain('🎉')
  })

  test('keeps a shortcode in an indented block literal', () => {
    const { container } = markdown('    print(":tada:")')
    expect(container.querySelector('code').textContent).toContain(':tada:')
  })

  test('leaves a link url alone but expands its text', () => {
    const { container } = markdown('[:star: here](https://example.com/:star:)')
    const link = container.querySelector('a')
    expect(link.textContent).toBe('⭐ here')
    expect(link).toHaveAttribute('href', 'https://example.com/:star:')
  })
})
