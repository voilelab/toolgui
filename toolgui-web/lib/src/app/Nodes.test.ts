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
