describe('Sidebar', () => {
  it('Sidebar test', () => {
    cy.visit('/sidebar')
    cy.get('div[id=container_component_container_sidebar]').should('not.exist')

    cy.contains('Show sidebar').click()
    cy.get('div[id=container_component_container_sidebar]').contains('Sidebar is here').should('exist')

    cy.contains('Show sidebar').click()
    cy.get('div[id=container_component_container_sidebar]').should('not.exist')
  })
})