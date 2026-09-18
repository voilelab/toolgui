// The app menubar: a tree declared once on the Go App, drawn as a row above
// the shell, whose items report a click to the run handling it.

describe('Menu', () => {
  // The item clicks turn scrolling off. Cypress scrolls what it is about to
  // click to the top of the viewport, and an item sits in a portal below the
  // entry it opened from -- scrolling it up carries the menubar off screen,
  // at which point Mantine hides the dropdown as detached from it. The
  // dropdown opens in view already, so there is nothing to scroll to.
  const noScroll = { scrollBehavior: false }

  // Opens a top level entry and returns its dropdown, which Mantine renders
  // in a portal rather than inside the row.
  function openMenu(label) {
    cy.get('.toolgui-menubar').contains('button', label).click()
    return cy.get('[role="menu"]:visible')
  }

  it('The menubar is a row above the shell', () => {
    cy.visit('/app_menu')

    cy.get('.toolgui-frame').should('have.class', 'has-menubar')
    cy.get('.toolgui-menubar').should('be.visible')
      .contains('button', 'File').should('exist')
    cy.get('.toolgui-menubar').contains('button', 'Help').should('exist')

    // Above the side column, not inside it.
    cy.get('.toolgui-menubar').then(($bar) => {
      cy.get('.toolgui-nav').then(($nav) => {
        expect($bar[0].getBoundingClientRect().bottom)
          .to.be.at.most($nav[0].getBoundingClientRect().top + 1)
      })
    })
    cy.get('.toolgui-nav .toolgui-menubar').should('not.exist')
  })

  // The row and the side column are pinned to the same edge, so scrolling
  // must not open a band of page between the column and the bottom of the
  // viewport -- which is what a row that scrolled away left behind.
  it('The menubar and the column stay put as the page scrolls', () => {
    cy.visit('/toolbar')

    cy.contains('row-39').scrollIntoView()

    cy.get('.toolgui-menubar').should($bar => {
      expect($bar[0].getBoundingClientRect().top).to.be.closeTo(0, 1)
    })

    cy.get('.toolgui-nav').should($nav => {
      const rect = $nav[0].getBoundingClientRect()
      expect(rect.bottom).to.be.closeTo(Cypress.config('viewportHeight'), 1)
    })
  })

  it('The menubar is on every page, whichever one is open', () => {
    cy.visit('/index')
    cy.get('.toolgui-menubar').contains('button', 'File').should('exist')

    cy.get('.toolgui-nav-list a[href="/layout"]').click()
    cy.get('.toolgui-menubar').contains('button', 'File').should('exist')
  })

  it('A submenu holds its items and its separator', () => {
    cy.visit('/app_menu')

    openMenu('File').within(() => {
      cy.contains('Say hello').should('exist')
      cy.contains('Clear the log').should('exist')
      cy.get('.mantine-Menu-divider').should('have.length', 1)
    })
  })

  // A menubar is armed by a click: crossing the row on the way somewhere else
  // must not pop a dropdown open, and once one is open, moving along the row
  // moves the dropdown rather than leaving two up.
  it('Opens one entry at a time, and only once armed', () => {
    cy.visit('/app_menu')

    cy.get('.toolgui-menubar').contains('button', 'File').trigger('mouseover')
    cy.get('[role="menu"]').should('not.exist')

    openMenu('File')
    cy.get('[role="menu"]:visible').should('have.length', 1)

    cy.get('.toolgui-menubar').contains('button', 'Help').trigger('mouseover')
    cy.get('[role="menu"]:visible').should('have.length', 1)
    cy.get('[role="menu"]:visible').contains('About').should('exist')
    cy.get('[role="menu"]:visible').contains('Say hello').should('not.exist')
  })

  it('An item reports its click to the run handling it', () => {
    cy.visit('/app_menu')
    cy.contains('Nothing picked yet.').should('exist')

    openMenu('File')
    cy.get('#menu_item_hello').click(noScroll)
    cy.contains('1. File > Say hello').should('exist')
    cy.contains('Nothing picked yet.').should('not.exist')

    openMenu('File')
    cy.get('#menu_item_hello').click(noScroll)
    cy.contains('2. File > Say hello').should('exist')

    // The click belongs to that one run: a rerun must not report it again.
    cy.contains('button', 'Rerun').click()
    cy.contains('2. File > Say hello').should('exist')
    cy.contains('3. File > Say hello').should('not.exist')
  })

  it('A second submenu has items of its own', () => {
    cy.visit('/app_menu')

    openMenu('Help')
    cy.get('#menu_item_about').click(noScroll)
    cy.contains('1. Help > About: toolgui').should('exist')

    openMenu('File')
    cy.get('#menu_item_clear').click(noScroll)
    cy.contains('Nothing picked yet.').should('exist')
  })
})
