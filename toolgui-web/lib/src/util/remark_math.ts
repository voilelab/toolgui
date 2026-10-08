import type { Root, Text } from 'mdast'
import { SKIP, visit } from 'unist-util-visit'

// remarkMathGuard turns an inline formula back into text when a space follows
// its opening `$` or precedes its closing one, so `$5 and $10` stays prose.
// Runs after remark-math, which has no such rule.
export function remarkMathGuard() {
  return (tree: Root, file: { value: unknown }) => {
    const src = String(file.value)
    visit(tree, 'inlineMath', (node, index, parent) => {
      const start = node.position?.start.offset
      const end = node.position?.end.offset
      if (!parent || index === undefined || start === undefined || end === undefined) {
        return
      }

      const raw = src.slice(start, end)
      if (!/^\$+\s|\s\$+$/.test(raw)) {
        return
      }

      const text: Text = { type: 'text', value: raw, position: node.position }
      parent.children[index] = text
      return SKIP
    })
  }
}
