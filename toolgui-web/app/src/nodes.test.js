import { expect, test, describe } from 'vitest'

import { Forest } from '@toolgui-web/lib/src/app/Nodes'

const MAIN = 'container_component_main'
const SIDEBAR = 'container_component_sidebar'

// runPage replays one page run the way App does: a copy per pack, beginRun
// before the creates, endRun after them.
function runPage(forest, comps, success = true) {
  var f = forest.swallowCopy()
  f.beginRun()

  for (const comp of comps) {
    f = f.swallowCopy()
    f.createNode(comp, comp.parent ?? MAIN)
  }

  f = f.swallowCopy()
  f.endRun(success)
  return f
}

const text = (t) => ({ name: 'text_component', id: `text_component_${t}`, text: t })

describe('Forest', () => {
  test('drops the nodes a run did not send', () => {
    var f = new Forest([MAIN, SIDEBAR])

    f = runPage(f, [text('a'), text('b')])
    expect(Object.keys(f.nodes)).toHaveLength(4)

    f = runPage(f, [text('a')])
    expect(new Set(Object.keys(f.nodes)))
      .toEqual(new Set([MAIN, SIDEBAR, 'text_component_a']))
    expect(f.nodes[MAIN].children.map(n => n.props.id)).toEqual(['text_component_a'])
  })

  test('does not grow when every run sends a fresh id', () => {
    var f = new Forest([MAIN, SIDEBAR])

    // A divider's id is random per run, so each run replaces the previous one.
    for (var run = 0; run < 5; run++) {
      f = runPage(f, [{ name: 'divider_component', id: `divider_component_${run}` }])
      expect(Object.keys(f.nodes)).toHaveLength(3)
      expect(f.nodes[MAIN].children).toHaveLength(1)
    }
  })

  test('keeps the root containers across runs', () => {
    var f = new Forest([MAIN, SIDEBAR])

    for (var run = 0; run < 3; run++) {
      f = runPage(f, [text('a')])
      expect(f.nodes[MAIN]).toBeDefined()
      expect(f.nodes[SIDEBAR]).toBeDefined()
    }
  })

  test('updates a node in place when its id is stable', () => {
    var f = new Forest([MAIN, SIDEBAR])

    f = runPage(f, [{ name: 'progress_bar_component', id: 'p', value: 1 }])
    const first = f.nodes['p']

    f = runPage(f, [{ name: 'progress_bar_component', id: 'p', value: 2 }])
    expect(f.nodes['p']).toBe(first)
    expect(f.nodes['p'].props.value).toBe(2)
  })

  test('keeps the page a failed run left behind', () => {
    var f = new Forest([MAIN, SIDEBAR])

    f = runPage(f, [text('a'), text('b')])
    // The page function panicked after the first component.
    f = runPage(f, [text('a')], false)

    expect(f.nodes[MAIN].children.map(n => n.props.id))
      .toEqual(['text_component_a', 'text_component_b'])
  })

  test('nests children under their own container', () => {
    var f = new Forest([MAIN, SIDEBAR])
    const box = { name: 'box_component', id: 'box' }
    const inner = { name: 'text_component', id: 'inner', text: 'x', parent: 'box' }

    f = runPage(f, [box, inner])
    expect(f.nodes[MAIN].children.map(n => n.props.id)).toEqual(['box'])
    expect(f.nodes['box'].children.map(n => n.props.id)).toEqual(['inner'])

    f = runPage(f, [box])
    expect(f.nodes['inner']).toBeUndefined()
    expect(f.nodes['box'].children).toHaveLength(0)
  })
})
