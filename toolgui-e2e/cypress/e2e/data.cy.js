describe('Data', () => {
  it('JSON test', () => {
    cy.visit('/data')
    cy.contains('"IsOk":').should('exist')
  })

  it('JSON expand test', () => {
    cy.visit('/data')
    cy.get('b').contains('{').click()
    cy.contains('{ ... }').should('exist')
  })

  it('Table test', () => {
    cy.visit('/data')
    cy.get('td').contains('2').should('exist')
  })

  // The canvas has no assertable content, so check that the component
  // mounted and that chart.js sized it.
  it('Chart test', () => {
    cy.visit('/data')

    cy.get('canvas#chart_component_demo_line')
      .should('have.attr', 'aria-label', 'line chart of visits, signups')

    cy.get('canvas#chart_component_demo_line').should(($canvas) => {
      expect($canvas[0].width).to.be.greaterThan(0)
    })

    cy.get('canvas#chart_component_demo_bar').should('exist')
    cy.get('canvas#chart_component_demo_area').should('exist')
  })

  it('Scatter chart test', () => {
    cy.visit('/data')

    cy.get('canvas#chart_component_demo_scatter')
      .should('have.attr', 'aria-label', 'scatter chart of runs')

    cy.get('canvas#chart_component_demo_scatter').should(($canvas) => {
      expect($canvas[0].width).to.be.greaterThan(0)
    })
  })
})

// The point of a DataFrame is that sorting, searching and paging happen in
// the browser. trackSends wraps the page's WebSocket so a test can assert
// that an interaction sent nothing to the server.
function trackSends(win) {
  const Native = win.WebSocket
  win.tgSent = []

  function Tracked(...args) {
    const conn = new Native(...args)
    const send = conn.send.bind(conn)
    conn.send = (data) => {
      win.tgSent.push(data)
      send(data)
    }
    return conn
  }

  Tracked.prototype = Native.prototype
  win.WebSocket = Tracked
}

const df = '#dataframe_component_demo_orders'

// amountOf reads the Amount cell, the last of the five columns, of the nth
// row of the current page.
function amountOf(n) {
  return cy.get(`${df} tbody tr`).eq(n).find('td').eq(4)
}

describe('DataFrame', () => {
  beforeEach(() => {
    cy.visit('/data', { onBeforeLoad: trackSends })
  })

  // PageSize is 10, so only ten of the two thousand rows are ever rendered.
  // That is the whole reason a big DataFrame stays smooth.
  it('renders one page of rows', () => {
    cy.get(`${df} tbody tr`).should('have.length', 10)
    amountOf(0).should('have.text', '1')
  })

  // A number column sorts by value: descending starts at 2000, where a
  // string sort would have started at "999".
  it('sorts a number column by value', () => {
    cy.get(`${df} thead th`).eq(4).find('button').click()
    amountOf(0).should('have.text', '1')

    cy.get(`${df} thead th`).eq(4).find('button').click()
    amountOf(0).should('have.text', '2000')
    amountOf(1).should('have.text', '1999')

    // A third click drops the sort and gives the rows back in server order.
    cy.get(`${df} thead th`).eq(4).find('button').click()
    amountOf(0).should('have.text', '1')
  })

  it('searches every column', () => {
    cy.get(`${df} input`).type('ORD-1500')
    cy.get(`${df} tbody tr`).should('have.length', 1)
    amountOf(0).should('have.text', '1500')

    // A hidden-from-view page of rows is still reachable by any cell.
    cy.get(`${df} input`).clear().type('LATAM')
    cy.get(`${df} tbody tr`).should('have.length', 10)
    amountOf(0).should('have.text', '3')
  })

  it('pages through the rows', () => {
    cy.get(`${df} .mantine-Pagination-control`).contains('2').click()
    amountOf(0).should('have.text', '11')
    cy.get(`${df} tbody tr`).should('have.length', 10)
  })

  it('shows a message when nothing matches', () => {
    cy.get(`${df} input`).type('no such order')
    cy.get(`${df} tbody tr`).should('have.length', 0)
    cy.get(df).contains('No rows').should('exist')
  })

  // The acceptance condition of the whole component: none of the above talks
  // to the server.
  it('does not talk to the server', () => {
    // Wait for the connection to settle, then count from there: the client
    // sends its state id when the socket opens.
    cy.get(`${df} tbody tr`).should('have.length', 10)

    cy.window().then((win) => {
      const before = win.tgSent.length

      cy.get(`${df} thead th`).eq(4).find('button').click()
      cy.get(`${df} input`).type('EMEA')
      cy.get(`${df} .mantine-Pagination-control`).contains('2').click()
      amountOf(0).should('have.text', '42')

      cy.window().its('tgSent').should('have.length', before)
    })
  })
})

const dfHosts = '#dataframe_component_demo_hosts'
const dfBuilds = '#dataframe_component_demo_builds'

// Picking a row is the one DataFrame interaction that reruns the page
// function, so unlike the block above, these read the answer off the page.
describe('DataFrame selection', () => {
  beforeEach(() => {
    cy.visit('/data', { onBeforeLoad: trackSends })
  })

  const selectedHosts = () => cy.get('#text_component_dataframe_multi_result')
  const selectedBuild = () => cy.get('#text_component_dataframe_single_result')

  // The checkbox column is column 0, so the Host cell is column 1. Clicking a
  // cell rather than the checkbox is the path a row click takes.
  const hostCell = (n) => cy.get(`${dfHosts} tbody tr`).eq(n).find('td').eq(1)
  const hostCheckbox = (n) =>
    cy.get(`${dfHosts} tbody tr`).eq(n).find('input[type=checkbox]')

  // The single-select table grows no checkbox column, so its first cell is
  // column 0.
  const buildCell = (n) => cy.get(`${dfBuilds} tbody tr`).eq(n).find('td').eq(0)

  // The search box is reached by its label: Mantine's TextInput carries no
  // type attribute, and a bare `input` would now also match the checkboxes.
  const hostSearch = `${dfHosts} input[aria-label="search the table"]`

  it('sends a picked row back to the page function', () => {
    selectedHosts().should('have.text', 'Selected: none')

    hostCheckbox(1).click()
    selectedHosts().should('have.text', 'Selected: web-2')

    // The result is in row order, not in the order the rows were picked.
    hostCheckbox(0).click()
    selectedHosts().should('have.text', 'Selected: web-1, web-2')
  })

  it('picks a row clicked anywhere, and drops it when clicked again', () => {
    hostCell(2).click()
    selectedHosts().should('have.text', 'Selected: db-1')
    cy.get(`${dfHosts} tbody tr`).eq(2)
      .should('have.attr', 'aria-selected', 'true')

    hostCell(2).click()
    selectedHosts().should('have.text', 'Selected: none')
  })

  // The index that goes back is the server's own, so a table the user has
  // reordered still names the row the page function wrote.
  it('picks the right row after sorting', () => {
    // Host ascending puts db-1 first, where the server sent it third.
    cy.get(`${dfHosts} thead th`).eq(1).find('button').click()
    hostCell(0).should('have.text', 'db-1')

    hostCheckbox(0).click()
    selectedHosts().should('have.text', 'Selected: db-1')
  })

  it('picks the right row after searching, and keeps it once the search clears', () => {
    cy.get(hostSearch).type('db-2')
    cy.get(`${dfHosts} tbody tr`).should('have.length', 1)

    hostCheckbox(0).click()
    selectedHosts().should('have.text', 'Selected: db-2')

    cy.get(hostSearch).clear()
    cy.get(`${dfHosts} tbody tr`).should('have.length', 4)
    hostCheckbox(3).should('be.checked')
  })

  it('takes every row the search kept from the head checkbox', () => {
    cy.get(hostSearch).type('web')
    cy.get(`${dfHosts} tbody tr`).should('have.length', 2)

    cy.get(`${dfHosts} thead input[type=checkbox]`).click()
    selectedHosts().should('have.text', 'Selected: web-1, web-2')

    // The rows the search filtered out were never taken, so clearing it
    // leaves them unpicked.
    cy.get(hostSearch).clear()
    cy.get(`${dfHosts} thead input[type=checkbox]`)
      .should('have.attr', 'data-indeterminate', 'true')
  })

  it('takes one row at a time in single mode', () => {
    // DefaultSelection names the first build, so the page starts on it.
    selectedBuild().should('have.text', 'Build: #41 / Go / passed')

    buildCell(1).click()
    selectedBuild().should('have.text', 'Build: #42 / Rust / failed')
    cy.get(`${dfBuilds} tbody tr`).eq(0)
      .should('have.attr', 'aria-selected', 'false')

    // Clicking the picked row again empties the selection.
    buildCell(1).click()
    selectedBuild().should('have.text', 'Build: none')
  })

  it('grows no checkbox column in single mode', () => {
    cy.get(`${dfBuilds} input[type=checkbox]`).should('not.exist')
    cy.get(`${dfBuilds} thead th`).should('have.length', 3)
  })

  // The counterpart of the DataFrame block's last test: this is the one
  // interaction that is meant to reach the server.
  it('talks to the server when a row is picked', () => {
    cy.get(`${dfHosts} tbody tr`).should('have.length', 4)

    cy.window().then((win) => {
      const before = win.tgSent.length

      hostCheckbox(0).click()
      selectedHosts().should('have.text', 'Selected: web-1')

      cy.window().its('tgSent').should('have.length.greaterThan', before)
    })
  })
})
