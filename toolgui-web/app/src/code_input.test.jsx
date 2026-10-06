import React from 'react'
import { act, cleanup, fireEvent, waitFor } from '@testing-library/react'
import { afterEach, expect, test, describe, vi } from 'vitest'

import { EditorView } from '@codemirror/view'
import { language } from '@codemirror/language'
import { openSearchPanel } from '@codemirror/search'

import { render } from './render'

import { Node } from '@toolgui-web/lib/src/app/Nodes'
import { TCodeInput } from '@toolgui-web/lib/src/components/tcinput/code_input'
import { clearState, stateValues } from '@toolgui-web/lib/src/components/state'

afterEach(() => {
  cleanup()
  clearState()
})

function codeNode(props) {
  return new Node('main/0', {
    name: 'code_input_component', id: 'code', label: 'Script',
    lang: 'go', height: 5, default: 'package main', reset_key: '',
    ...props,
  })
}

// The editor is lazy, so it shows up a tick after the render.
async function editorIn(container) {
  const content = await waitFor(() => {
    const el = container.querySelector('#code')
    expect(el).not.toBeNull()
    return el
  })
  return { content, view: EditorView.findFromDOM(content) }
}

async function renderCode(props, update = vi.fn()) {
  const r = render(
    <TCodeInput node={codeNode(props)} update={update} theme="light" />)
  return { ...r, ...await editorIn(r.container), update }
}

function type(view, text) {
  act(() => {
    view.dispatch({
      changes: { from: view.state.doc.length, insert: text },
    })
  })
}

describe('TCodeInput', () => {
  test('shows the default under its label', async () => {
    const { content, view, getByText } = await renderCode()

    expect(view.state.doc.toString()).toBe('package main')
    const label = getByText('Script')
    expect(content).toHaveAttribute('aria-labelledby', label.id)
  })

  test('starts on the value the app user typed before', async () => {
    stateValues.code = 'typed'
    const { view } = await renderCode()

    expect(view.state.doc.toString()).toBe('typed')
  })

  test('sends the code on focus out, once per change', async () => {
    const { content, view, update } = await renderCode()

    fireEvent.focusOut(content)
    expect(update).not.toHaveBeenCalled()

    type(view, '\n')
    expect(stateValues.code).toBe('package main\n')

    fireEvent.focusOut(content)
    fireEvent.focusOut(content)
    expect(update).toHaveBeenCalledTimes(1)
    expect(update).toHaveBeenCalledWith(
      { type: 'input', id: 'code', value: 'package main\n' })
  })

  test('keeps the code while focus moves to its search panel', async () => {
    const { content, view, update } = await renderCode()

    type(view, '\n')
    act(() => { openSearchPanel(view) })
    const search = view.dom.querySelector('.cm-search input')
    fireEvent.focusOut(content, { relatedTarget: search })
    expect(update).not.toHaveBeenCalled()

    fireEvent.focusOut(search)
    expect(update).toHaveBeenCalledWith(
      { type: 'input', id: 'code', value: 'package main\n' })
  })

  test('sends the code on Mod-Enter', async () => {
    const { content, view, update } = await renderCode()

    type(view, '\n')
    fireEvent.keyDown(content, { key: 'Enter', keyCode: 13, ctrlKey: true })
    fireEvent.keyDown(content, { key: 'Enter', keyCode: 13, metaKey: true })

    expect(update).toHaveBeenCalledTimes(1)
    expect(update).toHaveBeenCalledWith(
      { type: 'input', id: 'code', value: 'package main\n' })
  })

  test('a new reset_key restores the default', async () => {
    const update = vi.fn()
    const { rerender, container, view: first } =
      await renderCode({ reset_key: 'a' }, update)
    type(first, '!')

    rerender(<TCodeInput node={codeNode({ reset_key: 'b' })}
      update={update} theme="light" />)

    const { view } = await editorIn(container)
    expect(view).not.toBe(first)
    expect(view.state.doc.toString()).toBe('package main')
  })

  test('loads the language to highlight', async () => {
    const { view } = await renderCode({ lang: 'py' })

    await waitFor(() =>
      expect(view.state.facet(language)?.name).toBe('python'))
  })

  test('leaves an unknown language as plain text', async () => {
    const { view } = await renderCode({ lang: 'cobol' })

    await new Promise((r) => setTimeout(r, 0))
    expect(view.state.facet(language)).toBeNull()
  })
})
