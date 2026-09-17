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

  // The demo declares a page per component, which is more than the column is
  // tall. The list takes the overflow so the parts under it -- the page's own
  // sidebar, the controls, the version line -- stay where they are.
  it('A long page list scrolls inside the column', () => {
    cy.visit('/index')

    cy.get('.toolgui-nav-list').then(([list]) => {
      expect(list.scrollHeight).to.be.greaterThan(list.clientHeight)
    })

    cy.get('.toolgui-nav-foot').should('be.visible')
    cy.get('.toolgui-nav-version').should('be.visible')

    // Scrolling the list reaches the pages past the fold without moving
    // anything else.
    cy.get('.toolgui-nav-list').scrollTo('bottom')
    cy.get('.toolgui-nav-list a').last().should('be.visible')
    cy.get('.toolgui-nav-foot').should('be.visible')
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

    // The list scrolls here as well, so a page far enough down it is reached
    // by scrolling rather than by the bar growing to hold every link.
    cy.get('.toolgui-nav-list').contains('Layout')
      .scrollIntoView().should('be.visible')
  })

  // The bar the burger opens is a menu, not the page list laid end to end:
  // it fits the screen it was tapped on, with the controls and the version
  // line under the list rather than a thousand pixels past the fold.
  it('The open bar fits the screen on a narrow viewport', () => {
    cy.viewport(420, 800)
    cy.visit('/index')
    cy.get('.toolgui-nav-burger').click()

    cy.get('.toolgui-nav-list').then(([list]) => {
      expect(list.scrollHeight).to.be.greaterThan(list.clientHeight)
    })

    cy.get('.toolgui-nav').invoke('outerHeight').should('be.lessThan', 800)
    cy.get('.toolgui-nav-foot').should('be.visible')
    cy.get('.toolgui-nav-version').should('be.visible')
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

  // Dragging the handle: a real pointer sequence, so pointer capture and the
  // drag it drives are exercised the way a visitor would.
  function dragResizer(fromX, toX) {
    cy.get('.toolgui-nav-resizer')
      .trigger('pointerdown', { eventConstructor: 'PointerEvent', pointerId: 1, button: 0, clientX: fromX })
      .trigger('pointermove', { eventConstructor: 'PointerEvent', pointerId: 1, clientX: toX })
      .trigger('pointerup', { eventConstructor: 'PointerEvent', pointerId: 1, clientX: toX })
  }

  it('Dragging the handle widens the column and narrows the page', () => {
    cy.visit('/index')
    cy.get('.toolgui-nav').invoke('outerWidth').should('eq', 240)

    cy.get('.toolgui-main').invoke('outerWidth').then((mainWidth) => {
      dragResizer(240, 360)

      cy.get('.toolgui-nav').invoke('outerWidth').should('eq', 360)
      // Not exactly the 120 the column took: the page reflows at the narrower
      // width, and a document scrollbar appearing costs it another ~15.
      cy.get('.toolgui-main').invoke('outerWidth')
        .should('be.lessThan', mainWidth - 100)
    })
  })

  it('Dragging the other way narrows the column', () => {
    cy.visit('/index')
    dragResizer(240, 200)
    cy.get('.toolgui-nav').invoke('outerWidth').should('eq', 200)
  })

  // Past either bound the column stops, rather than squeezing the page list
  // out or eating the page.
  it('The drag stops at the bounds', () => {
    cy.visit('/index')
    dragResizer(240, 40)
    cy.get('.toolgui-nav').invoke('outerWidth').should('eq', 180)

    dragResizer(180, 1200)
    cy.get('.toolgui-nav').invoke('outerWidth').should('eq', 480)
  })

  // Same as the collapsed state: only what was stored survives the reload
  // jumpToPage does.
  it('The width survives a page change, and a double click resets it', () => {
    cy.visit('/index')
    dragResizer(240, 360)

    cy.visit('/layout')
    cy.get('.toolgui-nav').invoke('outerWidth').should('eq', 360)

    cy.get('.toolgui-nav-resizer').dblclick()
    cy.get('.toolgui-nav').invoke('outerWidth').should('eq', 240)
    cy.reload()
    cy.get('.toolgui-nav').invoke('outerWidth').should('eq', 240)
  })

  // The handle is a separator that Tab reaches and the arrows move.
  it('The handle resizes from the keyboard', () => {
    cy.visit('/index')
    cy.get('.toolgui-nav-resizer')
      .should('have.attr', 'role', 'separator')
      .should('have.attr', 'aria-valuenow', '240')
      .focus().should('have.focus')
      .trigger('keydown', { key: 'ArrowRight' })
      .trigger('keydown', { key: 'ArrowRight' })

    cy.get('.toolgui-nav').invoke('outerWidth').should('eq', 272)
    cy.get('.toolgui-nav-resizer').should('have.attr', 'aria-valuenow', '272')

    cy.get('.toolgui-nav-resizer').trigger('keydown', { key: 'End' })
    cy.get('.toolgui-nav').invoke('outerWidth').should('eq', 480)
  })

  // A first visit has nothing of the visitor's stored to size the column by.
  it('A first visit lands on the default width', () => {
    cy.visit('/index')
    cy.get('.toolgui-nav').invoke('outerWidth').should('eq', 240)
    cy.window().then((win) => {
      expect(win.localStorage.getItem('sidenav_width')).to.be.null
    })
  })

  // Collapsed there is no edge to move, and the burger owns a narrow
  // viewport, where the column is a top bar the full width of the screen.
  it('The handle is gone where the width is not the column to keep', () => {
    cy.visit('/index')
    cy.get('.toolgui-nav-resizer').should('be.visible')

    cy.get('.toolgui-nav-collapse').click()
    cy.get('.toolgui-nav-resizer').should('not.be.visible')

    cy.get('.toolgui-nav-collapse').click()
    cy.get('.toolgui-nav-resizer').should('be.visible')

    cy.viewport(420, 800)
    cy.get('.toolgui-nav-resizer').should('not.be.visible')
  })

  // A build from a tag reports it; anything else reports the pseudo-version
  // the toolchain derives from the commit, so only the shape is asserted.
  it('The column shows the toolgui version', () => {
    cy.visit('/index')
    cy.get('.toolgui-nav-version').invoke('text')
      .should('match', /^\s*toolgui v\S+\s*$/)
  })
})
