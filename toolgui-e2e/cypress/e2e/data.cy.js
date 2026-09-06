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
})