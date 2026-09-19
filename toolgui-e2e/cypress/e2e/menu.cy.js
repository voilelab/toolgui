// The app menubar: a tree declared once on the Go App, drawn as a row above
// the shell, whose items report a click to the run handling it.

describe('Menu', () => {
  // The item clicks turn scrolling off. Cypress scrolls what it is about to
  // click to the top of the viewport, which is where the menubar is pinned:
  // an item scrolled up there lands under the row that opened it. The
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

  // The browser's half: the shell listens for the keystroke and sends the
  // same click the item's own would have. On the desktop the OS dispatches it
  // off the native menu item, which is not what runs here.
  describe('Accelerators', () => {
    // Cypress types into the focused element; the body is what has focus
    // until something on the page takes it.
    function press(combo) {
      cy.get('body').type(combo)
    }

    it('An accelerator fires the item it was declared on', () => {
      cy.visit('/app_menu')
      cy.contains('Nothing picked yet.').should('exist')

      press('{ctrl}e')
      cy.contains('1. File > Say hello').should('exist')

      // Down a nested submenu, and a different combination for a different
      // item: the modifiers are matched, not just the key.
      press('{ctrl}{shift}E')
      cy.contains('2. FILE > MORE > SAY HELLO LOUDLY').should('exist')

      press('{ctrl}{shift}{backspace}')
      cy.contains('Nothing picked yet.').should('exist')
    })

    it('An item shows the combination that fires it', () => {
      cy.visit('/app_menu')

      openMenu('File').within(() => {
        cy.contains('.toolgui-menubar-accel', 'Ctrl+E').should('exist')
      })
    })

    // A bare key is a shortcut on the page and a keystroke in a field. F2 is
    // Help > About, and the page has a textbox to prove it with.
    //
    // A function key is dispatched rather than typed: cy.type has a sequence
    // for the modifiers and for Backspace, but none for F1 to F24.
    //
    // eventConstructor, because cy.trigger builds a plain Event by default
    // and assigns the options onto it -- which leaves ctrlKey and the rest
    // undefined rather than false, and the shell matches a keystroke by every
    // modifier, the ones the item did not ask for included.
    it('A bare key stays out of a text field', () => {
      const f2 = ['keydown',
        { eventConstructor: 'KeyboardEvent', key: 'F2', code: 'F2' }]
      const textbox = 'input[id=textbox_component_menu_typing]'

      cy.visit('/app_menu')

      cy.get('body').trigger(...f2)
      cy.contains('1. Help > About: toolgui').should('exist')

      cy.get(textbox).type('hello')
      cy.get(textbox).trigger(...f2)
      cy.contains('2. Help > About: toolgui').should('not.exist')
      cy.get(textbox).should('have.value', 'hello')

      // A real chord still reaches the menu from in there.
      cy.get(textbox).type('{ctrl}e')
      cy.contains('2. File > Say hello').should('exist')
    })
  })
})
