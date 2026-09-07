describe('Misc', () => {
  // The message is shown in an alert, not dropped into the page as text.
  it('Error handling', () => {
    cy.visit('/misc')
    cy.contains('Show error').click()
    cy.get('[role=alert]').contains('new error').should('exist')
  })

  it('Panic handling', () => {
    cy.visit('/misc')
    cy.contains('Show panic').click()
    cy.get('[role=alert]').contains('show panic').should('exist')
  })

  it('Html component', () => {
    cy.visit('/misc')
    cy.get('b').contains('Hello world gen by html component').should('exist')
  })
})