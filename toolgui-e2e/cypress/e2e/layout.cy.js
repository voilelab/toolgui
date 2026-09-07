describe('Layout spec', () => {
  it('Column works', () => {
    cy.visit('/layout')
    cy.contains('col-0').should('exist')
    cy.contains('col-1').should('exist')
    cy.contains('col-2').should('exist')
  })

  it('Box works', () => {
    cy.visit('/layout')
    cy.get('.box').contains('A box!').should('exist')
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
    cy.get('.tabs').next().contains('A tab!').invoke('attr', 'data-e2e', 'tab1')

    cy.get('.tabs').contains('tab2').click()
    cy.get('[data-e2e=tab1]').should('exist').and('not.be.visible')

    cy.get('.tabs').contains('tab1').click()
    cy.get('[data-e2e=tab1]').should('be.visible')
  })

  it('Expand works', () => {
    cy.visit('/layout')
    cy.contains('Expand').should('exist')
  })
})