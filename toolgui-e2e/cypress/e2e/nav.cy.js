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

  it('Page links are real links', () => {
    cy.visit('/index')
    cy.get('.toolgui-nav .menu-list').contains('Layout')
      .should('have.attr', 'href', '/layout')
      .focus().should('have.focus')
  })

  it('The navigation landmark covers only the page list', () => {
    cy.visit('/sidebar')
    cy.contains('Show sidebar').click()
    cy.get('.toolgui-nav').contains('Sidebar is here').should('exist')

    cy.get('.toolgui-nav').should('not.have.attr', 'role')
    cy.get('nav[aria-label="main navigation"]').within(() => {
      cy.get('.menu-list').should('exist')
      cy.contains('Sidebar is here').should('not.exist')
      cy.contains('Rerun').should('not.exist')
    })
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

  // A visitor with nothing stored still gets a theme, taken from the browser
  // preference Cypress runs with.
  it('A first visit lands on a real theme', () => {
    cy.visit('/index')
    cy.get('html').should('have.class', 'theme-light')
    cy.window().then((win) => {
      expect(win.localStorage.getItem('theme_mode')).to.be.null
    })
  })

  it('The theme toggle switches the theme and remembers it', () => {
    cy.visit('/index')
    cy.get('.toolgui-nav-foot .button').last().click()
    cy.get('html').should('have.class', 'theme-dark')

    cy.reload()
    cy.get('html').should('have.class', 'theme-dark')
  })

  // A build from a tag reports it; anything else reports the pseudo-version
  // the toolchain derives from the commit, so only the shape is asserted.
  it('The column shows the toolgui version', () => {
    cy.visit('/index')
    cy.get('.toolgui-nav-version').invoke('text')
      .should('match', /^\s*toolgui v\S+\s*$/)
  })
})
