import { expect, test, describe } from 'vitest'

import { Forest } from '@toolgui-web/lib/src/app/Nodes'

const MAIN = 'container_component_main'
const SIDEBAR = 'container_component_sidebar'

// runPage replays one page run: a forest copy per pack the way App does, and
// keys handed out by a per-container counter the way Container does.
function runPage(forest, build, success = true) {
  var f = forest.swallowCopy()
  f.beginRun()

  const counters = {}
  const add = (parentKey, props) => {
    const index = counters[parentKey] ?? 0
    counters[parentKey] = index + 1
    const key = `${parentKey}/${index}`

    f = f.swallowCopy()
    f.createNode(key, parentKey, index, props)
    return key
  }

  build(add)

  f = f.swallowCopy()
  f.endRun(success)
  return f
}

const text = (t) => ({ name: 'text_component', id: '', text: t })
const box = (id) => ({ name: 'container_component', id })
const widget = (id) => ({ name: 'textbox_component', id })

const shown = (f, key = MAIN) => f.nodes[key].children.map(n => n.props.text ?? n.props.id)

describe('Forest', () => {
  test('renders every copy of an identical component', () => {
    var f = new Forest([MAIN, SIDEBAR])

    f = runPage(f, add => {
      add(MAIN, text('duplicate me'))
      add(MAIN, text('duplicate me'))
    })

    expect(shown(f)).toEqual(['duplicate me', 'duplicate me'])
  })

  test('drops the nodes a run did not send', () => {
    var f = new Forest([MAIN, SIDEBAR])

    f = runPage(f, add => { add(MAIN, text('a')); add(MAIN, text('b')) })
    expect(Object.keys(f.nodes)).toHaveLength(4)

    f = runPage(f, add => { add(MAIN, text('a')) })
    expect(new Set(Object.keys(f.nodes)))
      .toEqual(new Set([MAIN, SIDEBAR, `${MAIN}/0`]))
    expect(shown(f)).toEqual(['a'])
  })

  test('keeps the order when a conditional drops a component in the middle', () => {
    var f = new Forest([MAIN, SIDEBAR])

    f = runPage(f, add => { add(MAIN, text('a')); add(MAIN, text('b')); add(MAIN, text('c')) })
    expect(shown(f)).toEqual(['a', 'b', 'c'])

    f = runPage(f, add => { add(MAIN, text('a')); add(MAIN, text('c')) })
    expect(shown(f)).toEqual(['a', 'c'])

    f = runPage(f, add => { add(MAIN, text('a')); add(MAIN, text('b')); add(MAIN, text('c')) })
    expect(shown(f)).toEqual(['a', 'b', 'c'])
  })

  test('does not grow across reruns', () => {
    var f = new Forest([MAIN, SIDEBAR])

    for (var run = 0; run < 5; run++) {
      f = runPage(f, add => {
        add(MAIN, { name: 'divider_component', id: '' })
        add(MAIN, { name: 'divider_component', id: '' })
      })
      expect(Object.keys(f.nodes)).toHaveLength(4)
      expect(f.nodes[MAIN].children).toHaveLength(2)
    }
  })

  test('keeps the root containers across runs', () => {
    var f = new Forest([MAIN, SIDEBAR])

    for (var run = 0; run < 3; run++) {
      f = runPage(f, add => { add(MAIN, text('a')) })
      expect(f.nodes[MAIN]).toBeDefined()
      expect(f.nodes[SIDEBAR]).toBeDefined()
    }
  })

  test('updates a node in place when it stays in the same place', () => {
    var f = new Forest([MAIN, SIDEBAR])

    f = runPage(f, add => { add(MAIN, { name: 'progress_bar_component', id: '', value: 1 }) })
    const first = f.nodes[`${MAIN}/0`]

    f = runPage(f, add => { add(MAIN, { name: 'progress_bar_component', id: '', value: 2 }) })
    expect(f.nodes[`${MAIN}/0`]).toBe(first)
    expect(f.nodes[`${MAIN}/0`].props.value).toBe(2)
  })

  test('replaces the node when the component type changes under it', () => {
    var f = new Forest([MAIN, SIDEBAR])

    f = runPage(f, add => { add(MAIN, text('a')) })
    const first = f.nodes[`${MAIN}/0`]

    f = runPage(f, add => { add(MAIN, { name: 'markdown_component', id: '', text: 'a' }) })
    expect(f.nodes[`${MAIN}/0`]).not.toBe(first)
  })

  test('keeps the page a failed run left behind', () => {
    var f = new Forest([MAIN, SIDEBAR])

    f = runPage(f, add => { add(MAIN, text('a')); add(MAIN, text('b')) })
    // The page function panicked after the first component.
    f = runPage(f, add => { add(MAIN, text('a')) }, false)

    expect(shown(f)).toEqual(['a', 'b'])
  })

  test('nests children under their own container', () => {
    var f = new Forest([MAIN, SIDEBAR])

    f = runPage(f, add => {
      const inner = add(MAIN, box('container_component_box_inner'))
      add(inner, text('x'))
    })
    expect(shown(f)).toEqual(['container_component_box_inner'])
    expect(shown(f, `${MAIN}/0`)).toEqual(['x'])

    f = runPage(f, add => { add(MAIN, box('container_component_box_inner')) })
    expect(f.nodes[`${MAIN}/0/0`]).toBeUndefined()
    expect(f.nodes[`${MAIN}/0`].children).toHaveLength(0)
  })

  test('removes a node the server deletes', () => {
    var f = new Forest([MAIN, SIDEBAR])

    f = runPage(f, add => { add(MAIN, text('a')); add(MAIN, text('b')) })
    f = f.swallowCopy()
    f.removeNode(`${MAIN}/0`)

    expect(shown(f)).toEqual(['b'])
    expect(f.nodes[`${MAIN}/0`]).toBeUndefined()
  })

  describe('reactKey', () => {
    test('is the id when the component has one', () => {
      var f = new Forest([MAIN, SIDEBAR])
      f = runPage(f, add => { add(MAIN, widget('textbox_component_Name')) })
      expect(f.nodes[`${MAIN}/0`].reactKey).toBe('textbox_component_Name')
    })

    test('follows a widget that a conditional moves', () => {
      var f = new Forest([MAIN, SIDEBAR])

      f = runPage(f, add => {
        add(MAIN, widget('textbox_component_A'))
        add(MAIN, widget('textbox_component_B'))
      })
      f = runPage(f, add => { add(MAIN, widget('textbox_component_B')) })

      // B moved from index 1 to index 0, but keeps the key it renders under,
      // so React does not hand it the state of the widget that was there.
      expect(f.nodes[`${MAIN}/0`].reactKey).toBe('textbox_component_B')
    })

    test('is the position and the type when the component has no id', () => {
      var f = new Forest([MAIN, SIDEBAR])
      f = runPage(f, add => { add(MAIN, text('a')); add(MAIN, text('a')) })

      expect(f.nodes[MAIN].children.map(n => n.reactKey))
        .toEqual([`${MAIN}/0:text_component`, `${MAIN}/1:text_component`])
    })

    test('changes when the component type at a position changes', () => {
      var f = new Forest([MAIN, SIDEBAR])
      f = runPage(f, add => { add(MAIN, text('a')) })
      const before = f.nodes[`${MAIN}/0`].reactKey

      f = runPage(f, add => { add(MAIN, { name: 'latex_component', id: '', latex: 'a' }) })

      // The renderer reaches every component through one TComponent, so a key
      // that stayed the same would carry the old component's hooks over.
      expect(f.nodes[`${MAIN}/0`].reactKey).not.toBe(before)
    })

    test('is never shared by two children while a run is in flight', () => {
      var f = new Forest([MAIN, SIDEBAR])
      f = runPage(f, add => {
        add(MAIN, widget('textbox_component_A'))
        add(MAIN, widget('textbox_component_B'))
      })

      // B moves up into A's slot. Until endRun the old B still sits in the
      // slot below, and rendering two children under one key is the collision
      // the whole scheme exists to avoid.
      var mid = f.swallowCopy()
      mid.beginRun()
      mid = mid.swallowCopy()
      mid.createNode(`${MAIN}/0`, MAIN, 0, widget('textbox_component_B'))

      const keys = mid.nodes[MAIN].children.filter(Boolean).map(n => n.reactKey)
      expect(new Set(keys).size).toBe(keys.length)
    })
  })
})
