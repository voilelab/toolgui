import { expect, test } from 'vitest'
import { unified } from 'unified'
import remarkParse from 'remark-parse'
import remarkMath from 'remark-math'
import type { Paragraph, Root } from 'mdast'

import { remarkMathGuard } from './remark_math'

function inline(md: string) {
  const processor = unified().use(remarkParse).use(remarkMath).use(remarkMathGuard)
  const tree = processor.runSync(processor.parse(md), md) as Root
  return (tree.children[0] as Paragraph).children
    .map(c => [c.type, 'value' in c ? c.value : ''])
}

test('keeps a formula', () => {
  expect(inline('x $a_1 + b$ y')).toEqual([
    ['text', 'x '], ['inlineMath', 'a_1 + b'], ['text', ' y']])
})

test('keeps prices as text', () => {
  expect(inline('costs $5 and $10 today')).toEqual([
    ['text', 'costs '], ['text', '$5 and $'], ['text', '10 today']])
})

test('rejects a space inside either dollar', () => {
  expect(inline('$ a$')).toEqual([['text', '$ a$']])
  expect(inline('$a $')).toEqual([['text', '$a $']])
})
