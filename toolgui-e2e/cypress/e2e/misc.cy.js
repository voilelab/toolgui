describe('Misc', () => {
  it('Error handling', () => {
    cy.visit('/misc')
    cy.contains('Show error').click()
    cy.contains('new error').should('exist')
  })

  it('Panic handling', () => {
    cy.visit('/misc')
    cy.contains('Show panic').click()
    cy.contains('show panic').should('exist')
  })

  it('Html component', () => {
    cy.visit('/misc')
    cy.get('b').contains('Hello world gen by html component').should('exist')
  })
})