import { describe, expect, it } from 'vitest'
import { Forest } from './Nodes'

const ROOT = 'container_component_main'

describe('keepNode', () => {
  it('keeps the same props reference across runs', () => {
    const f = new Forest([ROOT])
    const props = { name: 'dataframe_component', rows: [[1, 2]] }

    f.beginRun()
    f.createNode(`${ROOT}/0`, ROOT, 0, props)
    f.endRun(true)

    f.beginRun()
    f.keepNode(`${ROOT}/0`, ROOT, 0)
    f.endRun(true)

    expect(f.nodes[`${ROOT}/0`].props).toBe(props)
    expect(f.nodes[ROOT].children.map(n => n.key)).toEqual([`${ROOT}/0`])
  })

  it('survives endRun only when kept', () => {
    const f = new Forest([ROOT])

    f.beginRun()
    f.createNode(`${ROOT}/0`, ROOT, 0, { name: 'text_component' })
    f.createNode(`${ROOT}/1`, ROOT, 1, { name: 'text_component' })
    f.endRun(true)

    f.beginRun()
    f.keepNode(`${ROOT}/0`, ROOT, 0)
    f.endRun(true)

    expect(Object.keys(f.nodes).sort()).toEqual([ROOT, `${ROOT}/0`])
  })

  it('keeps the subtree of a kept container', () => {
    const f = new Forest([ROOT])
    const box = `${ROOT}/0`

    f.beginRun()
    f.createNode(box, ROOT, 0, { name: 'container_component', id: 'box' })
    f.createNode(`${box}/0`, box, 0, { name: 'text_component' })
    f.endRun(true)

    f.beginRun()
    f.keepNode(box, ROOT, 0)
    f.keepNode(`${box}/0`, box, 0)
    f.endRun(true)

    expect(f.nodes[box].children.map(n => n.key)).toEqual([`${box}/0`])
  })

  it('ignores a key it does not have', () => {
    const f = new Forest([ROOT])
    f.beginRun()
    f.keepNode(`${ROOT}/0`, ROOT, 0)
    expect(f.nodes[`${ROOT}/0`]).toBeUndefined()
  })
})

describe('node references', () => {
  const box = `${ROOT}/0`

  // A box holding two texts, then a box beside it.
  function build() {
    const f = new Forest([ROOT])
    f.beginRun()
    f.createNode(box, ROOT, 0, { name: 'box_component' })
    f.createNode(`${box}/0`, box, 0, { name: 'text_component', text: 'a' })
    f.createNode(`${box}/1`, box, 1, { name: 'text_component', text: 'b' })
    f.createNode(`${ROOT}/1`, ROOT, 1, { name: 'text_component', text: 'c' })
    f.endRun(true)
    return f
  }

  it('replaces a changed node and its ancestors only', () => {
    const before = build()
    const after = before.swallowCopy()
    after.updateNode(`${box}/1`, { name: 'text_component', text: 'B' })

    expect(after.nodes[`${box}/1`]).not.toBe(before.nodes[`${box}/1`])
    expect(after.nodes[box]).not.toBe(before.nodes[box])
    expect(after.nodes[ROOT]).not.toBe(before.nodes[ROOT])
    expect(after.nodes[`${box}/0`]).toBe(before.nodes[`${box}/0`])
    expect(after.nodes[`${ROOT}/1`]).toBe(before.nodes[`${ROOT}/1`])

    // The earlier copy is left as it was.
    expect(before.nodes[`${box}/1`].props.text).toBe('b')
    expect(after.nodes[box].children[1]).toBe(after.nodes[`${box}/1`])
    expect(after.nodes[ROOT].children[0]).toBe(after.nodes[box])
  })

  it('keeps every reference across a run of keeps', () => {
    let f = build()
    const before = { ...f.nodes }

    const step = (fn: (f: Forest) => void) => {
      f = f.swallowCopy()
      fn(f)
    }
    step(f => f.beginRun())
    step(f => f.keepNode(box, ROOT, 0))
    step(f => f.keepNode(`${box}/0`, box, 0))
    step(f => f.keepNode(`${box}/1`, box, 1))
    step(f => f.keepNode(`${ROOT}/1`, ROOT, 1))
    step(f => f.endRun(true))

    for (const key of Object.keys(before)) {
      expect(f.nodes[key]).toBe(before[key])
    }
  })

  it('replaces the parent of a removed node', () => {
    const before = build()
    const after = before.swallowCopy()
    after.removeNode(`${box}/0`)

    expect(after.nodes[box]).not.toBe(before.nodes[box])
    expect(after.nodes[box].children.map(n => n.key)).toEqual([`${box}/1`])
    expect(before.nodes[box].children).toHaveLength(2)
    expect(after.nodes[`${ROOT}/1`]).toBe(before.nodes[`${ROOT}/1`])
  })

  it('replaces only the containers a shorter run shrank', () => {
    const before = build()
    let f = before.swallowCopy()
    f.beginRun()
    f.keepNode(box, ROOT, 0)
    f.keepNode(`${box}/0`, box, 0)
    f.keepNode(`${ROOT}/1`, ROOT, 1)
    f = f.swallowCopy()
    f.endRun(true)

    expect(f.nodes[box]).not.toBe(before.nodes[box])
    expect(f.nodes[box].children.map(n => n.key)).toEqual([`${box}/0`])
    expect(f.nodes[`${box}/0`]).toBe(before.nodes[`${box}/0`])
    expect(f.nodes[`${ROOT}/1`]).toBe(before.nodes[`${ROOT}/1`])
    expect(f.nodes[ROOT].children[0]).toBe(f.nodes[box])
  })
})
