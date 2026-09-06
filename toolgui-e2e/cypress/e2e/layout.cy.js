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

  it('Expand works', () => {
    cy.visit('/layout')
    cy.contains('Expand').should('exist')
  })
})