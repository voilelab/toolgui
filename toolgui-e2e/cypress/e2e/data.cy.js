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
