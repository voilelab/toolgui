import React from 'react'
import { cleanup, screen, fireEvent } from '@testing-library/react'

import { render } from './render'
import { afterEach, expect, test, describe, vi } from 'vitest'

import { Node } from '@toolgui-web/lib/src/app/Nodes'
import { TDataFrame } from '@toolgui-web/lib/src/components/tcdata/dataframe'

const RENDER_PROPS = { update: vi.fn(), upload: vi.fn(), theme: 'light' }

// column is one entry of the columns prop, which the server sends with every
// default already settled.
const column = (type = 'text') => (
  { type, align: type === 'number' ? 'right' : 'left', width: '', hidden: false })

function mount({ head, rows, columns, pageSize = 25, ...rest }) {
  return render(
    <TDataFrame
      node={new Node('main/0', {
        name: 'dataframe_component',
        id: '',
        head,
        rows,
        columns,
        sortable: true,
        searchable: true,
        page_size: pageSize,
        height: '',
        ...rest,
      })}
      {...RENDER_PROPS} />
  )
}

// cells reads the cells the table draws, top to bottom. Queried rather than
// got, so that an empty table reads as [] instead of throwing.
const cells = () => screen.queryAllByRole('cell').map(c => c.textContent)

const sortBy = (label) => fireEvent.click(screen.getByRole('button', { name: `sort by ${label}` }))

// Vitest runs without globals, so RTL's auto-cleanup never registers.
afterEach(cleanup)

describe('TDataFrame sorting', () => {
  const mountNumbers = (values) => mount({
    head: ['n'],
    rows: values.map(v => [v]),
    columns: [column('number')],
  })

  test('sorts a number column by value, not as a string', () => {
    mountNumbers(['9', '10', '1'])

    sortBy('n')
    expect(cells()).toEqual(['1', '9', '10'])

    sortBy('n')
    expect(cells()).toEqual(['10', '9', '1'])
  })

  test('a third click gives the rows back in server order', () => {
    mountNumbers(['9', '10', '1'])

    sortBy('n')
    sortBy('n')
    sortBy('n')

    expect(cells()).toEqual(['9', '10', '1'])
  })

  // Descending is not ascending reversed: a cell that does not parse as its
  // column's type sorts last either way.
  test('keeps unparseable cells last in both directions', () => {
    mountNumbers(['9', '-', '10', ''])

    sortBy('n')
    expect(cells()).toEqual(['9', '10', '-', ''])

    sortBy('n')
    expect(cells()).toEqual(['10', '9', '-', ''])
  })

  test('keeps unparseable datetimes last in both directions', () => {
    mount({
      head: ['when'],
      rows: [['2026-01-02T00:00:00Z'], ['not a date'], ['2026-01-01T00:00:00Z']],
      columns: [column('datetime')],
    })

    sortBy('when')
    expect(cells()).toEqual(
      ['2026-01-01T00:00:00Z', '2026-01-02T00:00:00Z', 'not a date'])

    sortBy('when')
    expect(cells()).toEqual(
      ['2026-01-02T00:00:00Z', '2026-01-01T00:00:00Z', 'not a date'])
  })
})

describe('TDataFrame searching', () => {
  const mountRows = () => mount({
    head: ['name', 'secret'],
    rows: [['ada', 'alpha'], ['grace', 'beta']],
    columns: [column(), { ...column(), hidden: true }],
  })

  test('keeps only the matching rows, case insensitively', () => {
    mountRows()

    fireEvent.change(screen.getByLabelText('search the table'),
      { target: { value: 'ADA' } })

    expect(cells()).toEqual(['ada'])
  })

  // A hidden column is still searched, so a row can be found by a value it
  // does not show.
  test('searches a hidden column', () => {
    mountRows()

    fireEvent.change(screen.getByLabelText('search the table'),
      { target: { value: 'beta' } })

    expect(cells()).toEqual(['grace'])
  })

  test('shows a message when nothing matches', () => {
    mountRows()

    fireEvent.change(screen.getByLabelText('search the table'),
      { target: { value: 'nobody' } })

    expect(cells()).toEqual([])
    expect(screen.getByText('No rows')).toBeVisible()
  })
})

describe('TDataFrame paging', () => {
  test('renders one page at a time', () => {
    mount({
      head: ['n'],
      rows: [['1'], ['2'], ['3'], ['4'], ['5']],
      columns: [column('number')],
      pageSize: 2,
    })

    expect(cells()).toEqual(['1', '2'])

    fireEvent.click(screen.getByRole('button', { name: '3' }))
    expect(cells()).toEqual(['5'])
  })

  // Nothing above reaches the server: the component never calls update.
  test('never calls update', () => {
    mount({
      head: ['n'],
      rows: [['1'], ['2'], ['3']],
      columns: [column('number')],
      pageSize: 2,
    })

    sortBy('n')
    fireEvent.change(screen.getByLabelText('search the table'),
      { target: { value: '1' } })

    expect(RENDER_PROPS.update).not.toHaveBeenCalled()
  })
})

describe('TDataFrame head', () => {
  test('drops a hidden column and keeps the rest', () => {
    mount({
      head: ['shown', 'gone'],
      rows: [['a', 'b']],
      columns: [column(), { ...column(), hidden: true }],
    })

    expect(screen.getAllByRole('columnheader').map(h => h.textContent))
      .toEqual(['shown'])
    expect(cells()).toEqual(['a'])
  })

  test('draws the head with no rows at all', () => {
    mount({ head: ['n'], rows: [], columns: [column('number')] })

    expect(screen.getAllByRole('columnheader').map(h => h.textContent))
      .toEqual(['n'])
    expect(screen.getByText('No rows')).toBeVisible()
  })
})
