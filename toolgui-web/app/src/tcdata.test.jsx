import React from 'react'
import { cleanup, createEvent, screen, fireEvent, within } from '@testing-library/react'

import { render } from './render'
import { afterEach, beforeEach, expect, test, describe, vi } from 'vitest'

import { Node } from '@toolgui-web/lib/src/app/Nodes'
import { clearState } from '@toolgui-web/lib/src/components/state'
import { TDataFrame } from '@toolgui-web/lib/src/components/tcdata/dataframe'

const RENDER_PROPS = { update: vi.fn(), upload: vi.fn(), theme: 'light' }

// column is one entry of the columns prop, which the server sends with every
// default already settled.
const column = (type = 'text') => (
  { type, align: type === 'number' ? 'right' : 'left', width: '', hidden: false })

// nodeFor builds the node a mount renders, so a test can hand the same
// position a new set of props the way a rerun does.
function nodeFor({ head, rows, columns, pageSize = 25, ...rest }) {
  return new Node('main/0', {
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
  })
}

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

describe('TDataFrame selection', () => {
  const mountHosts = (rest) => mount({
    head: ['host', 'region'],
    rows: [['web-1', 'APAC'], ['web-2', 'EMEA'], ['db-1', 'NA']],
    columns: [column(), column()],
    id: 'hosts',
    selection: 'multi',
    default_selection: [],
    ...rest,
  })

  // stateValues outlives a test, so a pick made in one would otherwise be
  // inherited by the next table mounted under the same id.
  beforeEach(() => {
    clearState()
    RENDER_PROPS.update.mockClear()
  })

  // bodyRow reads the nth row on screen, which is not the nth row the server
  // sent once the table has been sorted or searched.
  const bodyRow = (n) => screen.getAllByRole('row')[n + 1]

  const checkboxAt = (n) => within(bodyRow(n)).getByRole('checkbox')

  // textAt skips the checkbox column, which is a cell holding no text.
  const textAt = (n) => within(bodyRow(n)).getAllByRole('cell')
    .map(c => c.textContent).filter(t => t !== '')

  // sent is the values of the last select event, which is what the page
  // function is about to be rerun with.
  const sent = () => {
    const calls = RENDER_PROPS.update.mock.calls
    return calls[calls.length - 1][0]
  }

  const search = (value) => fireEvent.change(
    screen.getByLabelText('search the table'), { target: { value } })

  test('sends the picked row as a select event', () => {
    mountHosts()

    fireEvent.click(checkboxAt(1))

    expect(RENDER_PROPS.update).toHaveBeenCalledTimes(1)
    expect(sent()).toEqual({ type: 'select', id: 'hosts', values: [1] })
  })

  // A row can be picked without aiming at its checkbox.
  test('picks a row clicked anywhere', () => {
    mountHosts()

    fireEvent.click(within(bodyRow(2)).getAllByRole('cell')[1])

    expect(sent().values).toEqual([2])
  })

  // The checkbox toggles the row once, not twice: its click must not also
  // reach the row underneath it.
  test('does not toggle twice when the checkbox itself is clicked', () => {
    mountHosts()

    fireEvent.click(checkboxAt(0))

    expect(RENDER_PROPS.update).toHaveBeenCalledTimes(1)
    expect(checkboxAt(0)).toBeChecked()
  })

  test('keeps the picks in row order whatever order they were made in', () => {
    mountHosts()

    fireEvent.click(checkboxAt(2))
    fireEvent.click(checkboxAt(0))

    expect(sent().values).toEqual([0, 2])
  })

  test('drops a row picked a second time', () => {
    mountHosts()

    fireEvent.click(checkboxAt(1))
    fireEvent.click(checkboxAt(1))

    expect(sent().values).toEqual([])
    expect(checkboxAt(1)).not.toBeChecked()
  })

  // The indices are the server's own, so a table the user has reordered still
  // names the row the page function wrote.
  test('sends the server row index, not the place on screen', () => {
    mountHosts()

    sortBy('host')
    expect(textAt(0)).toEqual(['db-1', 'NA'])

    fireEvent.click(checkboxAt(0))
    expect(sent().values).toEqual([2])
  })

  test('sends the server row index after a search', () => {
    mountHosts()

    search('db')
    expect(textAt(0)).toEqual(['db-1', 'NA'])

    fireEvent.click(checkboxAt(0))
    expect(sent().values).toEqual([2])
  })

  // A pick is a property of the row, so putting the row back on screen finds
  // it still picked.
  test('keeps a pick across a search that hides the row', () => {
    mountHosts()

    fireEvent.click(checkboxAt(2))
    search('web')
    expect(screen.getAllByRole('row')).toHaveLength(3)

    search('')
    expect(checkboxAt(2)).toBeChecked()
  })

  // The head checkbox covers every row the search kept, on whatever page they
  // sit, and not the rows it filtered out.
  test('the head checkbox takes every row the search kept', () => {
    mountHosts()

    search('web')
    fireEvent.click(screen.getByLabelText('select every row'))

    expect(sent().values).toEqual([0, 1])
  })

  test('the head checkbox gives back only what it took', () => {
    mountHosts()

    fireEvent.click(checkboxAt(2))
    search('web')
    fireEvent.click(screen.getByLabelText('select every row'))
    search('')

    // db-1 was picked by hand before the search, so clearing the web rows
    // leaves it alone.
    expect(sent().values).toEqual([0, 1, 2])

    fireEvent.click(screen.getByLabelText('select every row'))
    expect(sent().values).toEqual([])
  })

  test('the head checkbox is indeterminate on a partial pick', () => {
    mountHosts()

    fireEvent.click(checkboxAt(0))
    expect(screen.getByLabelText('select every row')).toBePartiallyChecked()

    fireEvent.click(checkboxAt(1))
    fireEvent.click(checkboxAt(2))
    expect(screen.getByLabelText('select every row')).toBeChecked()
  })

  // The checkbox is already a tab stop and already keyboard-operable, so the
  // row is not made a second one.
  test('leaves the row out of the tab order in multi mode', () => {
    mountHosts()

    expect(bodyRow(0)).not.toHaveAttribute('tabindex')
    expect(checkboxAt(0)).toBeInTheDocument()
  })

  test('shows the default before anything is touched', () => {
    mountHosts({ default_selection: [1] })

    expect(checkboxAt(1)).toBeChecked()
    expect(bodyRow(1)).toHaveAttribute('aria-selected', 'true')
    expect(bodyRow(0)).toHaveAttribute('aria-selected', 'false')
    expect(RENDER_PROPS.update).not.toHaveBeenCalled()
  })

  describe('single', () => {
    const mountSingle = (rest) => mountHosts({ selection: 'single', ...rest })

    test('grows no checkbox column', () => {
      mountSingle()

      expect(screen.queryAllByRole('checkbox')).toHaveLength(0)
      expect(screen.getAllByRole('columnheader')).toHaveLength(2)
    })

    test('replaces the pick rather than adding to it', () => {
      mountSingle()

      fireEvent.click(bodyRow(0))
      expect(sent().values).toEqual([0])

      fireEvent.click(bodyRow(2))
      expect(sent().values).toEqual([2])
      expect(bodyRow(0)).toHaveAttribute('aria-selected', 'false')
    })

    test('clears the pick when the picked row is clicked again', () => {
      mountSingle()

      fireEvent.click(bodyRow(1))
      fireEvent.click(bodyRow(1))

      expect(sent().values).toEqual([])
      expect(bodyRow(1)).toHaveAttribute('aria-selected', 'false')
    })

    // The row is the only control a single-select table has, so it has to be
    // reachable and operable without a mouse.
    test('makes every row a tab stop', () => {
      mountSingle()

      expect(bodyRow(0)).toHaveAttribute('tabindex', '0')
      expect(bodyRow(2)).toHaveAttribute('tabindex', '0')
    })

    test.each([['Enter'], [' ']])('picks the focused row on %s', (key) => {
      mountSingle()

      fireEvent.keyDown(bodyRow(2), { key })

      expect(sent().values).toEqual([2])
      expect(bodyRow(2)).toHaveAttribute('aria-selected', 'true')
    })

    test('clears the pick when the picked row is keyed again', () => {
      mountSingle()

      fireEvent.keyDown(bodyRow(1), { key: 'Enter' })
      fireEvent.keyDown(bodyRow(1), { key: 'Enter' })

      expect(sent().values).toEqual([])
    })

    // Space scrolls the page and Enter submits a surrounding form, so both
    // have to be taken rather than merely acted on.
    test('takes the key it acts on', () => {
      mountSingle()

      const event = createEvent.keyDown(bodyRow(0), { key: 'Enter' })
      fireEvent(bodyRow(0), event)

      expect(event.defaultPrevented).toBe(true)
    })

    test('leaves other keys alone', () => {
      mountSingle()

      const event = createEvent.keyDown(bodyRow(0), { key: 'a' })
      fireEvent(bodyRow(0), event)

      expect(event.defaultPrevented).toBe(false)
      expect(RENDER_PROPS.update).not.toHaveBeenCalled()
    })

    // A default naming more than one row is the server's to trim, but the
    // table shows whatever it is given rather than second-guessing it.
    test('shows the default', () => {
      mountSingle({ default_selection: [2] })

      expect(bodyRow(2)).toHaveAttribute('aria-selected', 'true')
    })
  })

  // The mode can change between runs while the component stays mounted, and
  // the hook initializer does not run again, so what is drawn has to follow
  // the mode rather than the selection alone.
  describe('a mode change between runs', () => {
    const props = (rest) => ({
      head: ['host', 'region'],
      rows: [['web-1', 'APAC'], ['web-2', 'EMEA'], ['db-1', 'NA']],
      columns: [column(), column()],
      id: 'hosts',
      selection: 'multi',
      default_selection: [],
      ...rest,
    })

    // rerenderAs replaces the props at the same position, which is what the
    // node tree does on a rerun -- the component is not remounted.
    const rerenderAs = (rerender, selection) => rerender(
      <TDataFrame node={nodeFor(props({ selection }))} {...RENDER_PROPS} />)

    const selectedRows = () => screen.getAllByRole('row').slice(1)
      .map(r => r.getAttribute('aria-selected'))

    test('trims a multi selection to one row under single', () => {
      const { rerender } = mount(props())

      fireEvent.click(checkboxAt(0))
      fireEvent.click(checkboxAt(2))
      expect(sent().values).toEqual([0, 2])

      // Go caps single at the lowest index, so the table must draw the same.
      rerenderAs(rerender, 'single')
      expect(selectedRows()).toEqual(['true', 'false', 'false'])
    })

    test('draws no row as picked once the rows are unpickable', () => {
      const { rerender } = mount(props())

      fireEvent.click(checkboxAt(0))
      fireEvent.click(checkboxAt(2))

      rerenderAs(rerender, 'none')

      // Neither the marker nor the highlight survives: an unpickable table
      // that still paints rows as selected is the worse half of this.
      expect(selectedRows()).toEqual([null, null, null])
      expect(screen.queryAllByRole('checkbox')).toHaveLength(0)
      screen.getAllByRole('row').slice(1).forEach(
        r => expect(r).not.toHaveStyle({ background: 'var(--mantine-color-blue-light)' }))
    })

    test('gives the rows back when the mode widens again', () => {
      const { rerender } = mount(props())

      fireEvent.click(checkboxAt(0))
      fireEvent.click(checkboxAt(2))

      rerenderAs(rerender, 'single')
      rerenderAs(rerender, 'multi')

      // The selection was only ever narrowed for drawing, never thrown away,
      // which is what the Go side does with the state too.
      expect(selectedRows()).toEqual(['true', 'false', 'true'])
    })
  })

  describe('none', () => {
    const mountNone = () => mountHosts({ selection: 'none', id: '' })

    test('draws no checkbox and marks no row selectable', () => {
      mountNone()

      expect(screen.queryAllByRole('checkbox')).toHaveLength(0)
      expect(bodyRow(0)).not.toHaveAttribute('aria-selected')
      expect(bodyRow(0)).not.toHaveAttribute('tabindex')
    })

    test('never calls update', () => {
      mountNone()

      fireEvent.click(bodyRow(0))
      fireEvent.click(within(bodyRow(1)).getAllByRole('cell')[0])

      expect(RENDER_PROPS.update).not.toHaveBeenCalled()
    })
  })
})
