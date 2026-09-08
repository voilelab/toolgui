describe('Nav', () => {
  it('Left column lists pages and marks the current one', () => {
    cy.visit('/index')
    cy.get('.toolgui-nav-list a[href="/index"]')
      .should('have.attr', 'aria-current', 'page')
    cy.get('.toolgui-nav-list a[href="/layout"]')
      .should('not.have.attr', 'aria-current')
  })

  it('Clicking a page navigates', () => {
    cy.visit('/index')
    cy.get('.toolgui-nav-list a[href="/layout"]').click()
    cy.location('pathname').should('eq', '/layout')
    cy.get('.toolgui-nav-list a[href="/layout"]')
      .should('have.attr', 'aria-current', 'page')
  })

  it('Page links are real links', () => {
    cy.visit('/index')
    cy.get('.toolgui-nav-list a').contains('Layout').closest('a')
      .should('have.attr', 'href', '/layout')
      .focus().should('have.focus')
  })

  it('The navigation landmark covers only the page list', () => {
    cy.visit('/sidebar')
    cy.contains('Show sidebar').click()
    cy.get('.toolgui-nav').contains('Sidebar is here').should('exist')

    cy.get('.toolgui-nav').should('not.have.attr', 'role')
    cy.get('nav[aria-label="main navigation"]').within(() => {
      cy.get('a[href="/index"]').should('exist')
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

  it('The column collapses and hands its width to the page', () => {
    cy.visit('/index')
    cy.get('.toolgui-nav').invoke('outerWidth').should('be.greaterThan', 200)

    cy.get('.toolgui-main').invoke('outerWidth').then((mainWidth) => {
      cy.get('.toolgui-nav-collapse').should('be.visible')
        .should('have.attr', 'aria-expanded', 'true')
        .should('have.attr', 'aria-controls', 'toolgui-nav-body')
        .click()

      cy.get('.toolgui-nav-body').should('not.be.visible')
      cy.get('.toolgui-nav').invoke('outerWidth').should('be.lessThan', 100)
      cy.get('.toolgui-main').invoke('outerWidth')
        .should('be.greaterThan', mainWidth + 100)
    })
  })

  // Collapsed leaves a handle, not a dead edge: still there, still a real
  // button, so Tab reaches it and Enter / Space fire it.
  it('The collapsed column keeps a reachable expand handle', () => {
    cy.visit('/index')
    cy.get('.toolgui-nav-collapse').click()

    cy.get('.toolgui-nav-collapse').should('be.visible')
      .should('have.attr', 'aria-expanded', 'false')
      .should('have.prop', 'tagName', 'BUTTON')
      .focus().should('have.focus')
      .click()

    cy.get('.toolgui-nav-body').should('be.visible')
    cy.get('.toolgui-nav-collapse').should('have.attr', 'aria-expanded', 'true')
  })

  // Hidden, not unmounted, so the components keep their values.
  it('Collapsing only hides the page sidebar', () => {
    cy.visit('/sidebar')
    cy.contains('Show sidebar').click()
    cy.get('div[id=container_component_container_sidebar]').should('be.visible')

    cy.get('.toolgui-nav-collapse').click()
    cy.get('div[id=container_component_container_sidebar]')
      .should('exist').should('not.be.visible')

    cy.get('.toolgui-nav-collapse').click()
    cy.get('div[id=container_component_container_sidebar]')
      .contains('Sidebar is here').should('be.visible')
  })

  // jumpToPage loads the next page from scratch, so component state is gone
  // by the time it paints; only what was stored survives. cy.visit is that
  // same load.
  it('The collapsed column stays collapsed across a page change', () => {
    cy.visit('/index')
    cy.get('.toolgui-nav-collapse').click()
    cy.get('.toolgui-nav-body').should('not.be.visible')

    cy.visit('/layout')
    cy.get('.toolgui-nav').should('have.class', 'is-collapsed')
    cy.get('.toolgui-nav-body').should('not.be.visible')
    cy.get('.toolgui-nav-collapse').should('have.attr', 'aria-expanded', 'false')

    // And expanding again is what the page after that comes back to.
    cy.get('.toolgui-nav-collapse').click()
    cy.reload()
    cy.get('.toolgui-nav-body').should('be.visible')
    cy.get('.toolgui-nav').should('not.have.class', 'is-collapsed')
  })

  // A visitor who has never touched the toggle gets the column expanded,
  // with nothing of theirs stored to say otherwise.
  it('A first visit lands on an expanded column', () => {
    cy.visit('/index')
    cy.get('.toolgui-nav-body').should('be.visible')
    cy.window().then((win) => {
      expect(win.localStorage.getItem('sidenav_collapsed')).to.be.null
    })
  })

  it('The column collapses behind a burger on a narrow viewport', () => {
    cy.viewport(420, 800)
    cy.visit('/index')
    cy.get('.toolgui-nav-body').should('not.be.visible')

    // The burger is the only toggle at this width.
    cy.get('.toolgui-nav-collapse').should('not.be.visible')

    cy.get('.toolgui-nav-burger').click()
    cy.get('.toolgui-nav-body').should('be.visible')
    cy.get('.toolgui-nav-body').contains('Layout').should('be.visible')
  })

  it('Rerun and theme controls stay reachable', () => {
    cy.visit('/index')
    cy.get('.toolgui-nav-foot').contains('Rerun').should('exist')
    cy.get('.toolgui-nav-foot button').should('have.length.at.least', 2)
  })

  // A visitor with nothing stored still gets a theme, taken from the browser
  // preference Cypress runs with. Mantine holds the theme, so the attribute
  // it stamps on <html> is what says which one is on.
  it('A first visit lands on a real theme', () => {
    cy.visit('/index')
    cy.get('html').should('have.attr', 'data-mantine-color-scheme', 'light')
    cy.window().then((win) => {
      expect(win.localStorage.getItem('theme_mode')).to.be.null
    })
  })

  it('The theme toggle switches the theme and remembers it', () => {
    cy.visit('/index')
    cy.get('.toolgui-nav-foot button').last().click()
    cy.get('html').should('have.attr', 'data-mantine-color-scheme', 'dark')

    cy.reload()
    cy.get('html').should('have.attr', 'data-mantine-color-scheme', 'dark')
  })

  // A build from a tag reports it; anything else reports the pseudo-version
  // the toolchain derives from the commit, so only the shape is asserted.
  it('The column shows the toolgui version', () => {
    cy.visit('/index')
    cy.get('.toolgui-nav-version').invoke('text')
      .should('match', /^\s*toolgui v\S+\s*$/)
  })
})
