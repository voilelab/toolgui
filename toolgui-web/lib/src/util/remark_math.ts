import type { Root, Text } from 'mdast'
import { SKIP, visit } from 'unist-util-visit'

// remarkMathGuard fixes up the inline formulas remark-math finds:
//   - `$$...$$` on one line is drawn as a block, as a multiline one is.
//   - `$...$` with a space after its opening `$` or before its closing one
//     goes back to text, so `$5 and $10` stays prose.
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
      if (raw.startsWith('$$')) {
        // rehype-katex draws a node in display mode by this class.
        node.data = {
          ...node.data,
          hProperties: { className: ['language-math', 'math-display'] },
        }
        return
      }
      if (!/^\$\s|\s\$$/.test(raw)) {
        return
      }

      const text: Text = { type: 'text', value: raw, position: node.position }
      parent.children[index] = text
      return SKIP
    })
  }
}
