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
})