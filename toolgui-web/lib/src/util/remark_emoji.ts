import type { Root } from 'mdast'
import { visit } from 'unist-util-visit'

import { emojize } from './emoji'

// remarkEmoji expands the emoji shortcodes in a markdown tree.
//
// It rewrites text nodes and nothing else. A code span and a fenced block
// are their own node types, so a shortcode inside one stays literal: there
// it is the thing being shown, not decoration. Same for a link's url and an
// image's src, which are node fields rather than text.
export function remarkEmoji() {
  return (tree: Root) => {
    visit(tree, 'text', (node) => {
      node.value = emojize(node.value)
    })
  }
}
