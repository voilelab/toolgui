describe('Layout spec', () => {
  it('Column works', () => {
    cy.visit('/layout')
    cy.contains('col-0').should('exist')
    cy.contains('col-1').should('exist')
    cy.contains('col-2').should('exist')
  })

  it('Box works', () => {
    cy.visit('/layout')
    cy.get('.toolgui-box').contains('A box!').should('exist')
  })

  it('Tab works', () => {
    cy.visit('/layout')
    cy.contains('tab1').click()
    cy.contains('A tab!').should('exist')

    cy.contains('tab2').click()
    cy.contains('B tab!').should('exist')
  })

  it('Tab keeps the tab you leave mounted', () => {
    cy.visit('/layout')

    // Stamp the node showing tab1's content; it should survive a round trip.
    cy.contains('[role=tabpanel]', 'A tab!').contains('A tab!')
      .invoke('attr', 'data-e2e', 'tab1')

    cy.contains('[role=tab]', 'tab2').click()
    cy.get('[data-e2e=tab1]').should('exist').and('not.be.visible')

    cy.contains('[role=tab]', 'tab1').click()
    cy.get('[data-e2e=tab1]').should('be.visible')
  })

  it('Expand works', () => {
    cy.visit('/layout')
    cy.contains('Expand').should('exist')
  })

  // The slot is written three times in one run, and what it holds at the end
  // is the only thing on the screen. Everything is scoped to the component
  // column: the column beside it shows the source, which names the same text.
  it('Empty writes over its contents instead of stacking them', () => {
    cy.visit('/layout')
    const shown = () => cy.get('#column_component_show_empty_0')

    shown().contains('No query yet.').should('exist')
    shown().contains('Run a slow query').click()
    shown().contains('Querying…').should('exist')

    shown().contains('orders', { timeout: 15000 }).should('exist')
    shown().contains('Querying…').should('not.exist')
    shown().contains('No query yet.').should('not.exist')
  })
})
