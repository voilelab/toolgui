describe('Nav', () => {
  it('Left column lists pages and marks the current one', () => {
    cy.visit('/index')
    cy.get('.toolgui-nav .menu-list').contains('Index')
      .should('have.class', 'is-active')
    cy.get('.toolgui-nav .menu-list').contains('Layout')
      .should('not.have.class', 'is-active')
  })

  it('Clicking a page navigates', () => {
    cy.visit('/index')
    cy.get('.toolgui-nav .menu-list').contains('Layout').click()
    cy.location('pathname').should('eq', '/layout')
    cy.get('.toolgui-nav .menu-list').contains('Layout')
      .should('have.class', 'is-active')
  })

  it('The page sidebar shares the left column', () => {
    cy.visit('/sidebar')
    cy.get('.toolgui-nav').contains('Sidebar is here').should('not.exist')

    cy.contains('Show sidebar').click()
    cy.get('.toolgui-nav').contains('Sidebar is here').should('exist')
  })

  it('The column collapses behind a burger on a narrow viewport', () => {
    cy.viewport(420, 800)
    cy.visit('/index')
    cy.get('.toolgui-nav-body').should('not.be.visible')

    cy.get('.toolgui-nav-burger').click()
    cy.get('.toolgui-nav-body').should('be.visible')
    cy.get('.toolgui-nav-body').contains('Layout').should('be.visible')
  })

  it('Rerun and theme controls stay reachable', () => {
    cy.visit('/index')
    cy.get('.toolgui-nav-foot').contains('Rerun').should('exist')
    cy.get('.toolgui-nav-foot .button').should('have.length.at.least', 2)
  })
})
