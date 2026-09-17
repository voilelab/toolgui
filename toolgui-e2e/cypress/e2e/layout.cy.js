describe('Layout spec', () => {
  it('Column works', () => {
    cy.visit('/layout')
    cy.contains('col-0').should('exist')
    cy.contains('col-1').should('exist')
    cy.contains('col-2').should('exist')
  })

  it('Box works', () => {
    cy.visit('/layout')
    cy.get('.toolgui-box').contains('A box!').should('exist')
  })

  it('Tab works', () => {
    cy.visit('/layout')
    cy.contains('tab1').click()
    cy.contains('A tab!').should('exist')

    cy.contains('tab2').click()
    cy.contains('B tab!').should('exist')
  })

  it('Tab keeps the tab you leave mounted', () => {
    cy.visit('/layout')

    // Stamp the node showing tab1's content; it should survive a round trip.
    cy.contains('[role=tabpanel]', 'A tab!').contains('A tab!')
      .invoke('attr', 'data-e2e', 'tab1')

    cy.contains('[role=tab]', 'tab2').click()
    cy.get('[data-e2e=tab1]').should('exist').and('not.be.visible')

    cy.contains('[role=tab]', 'tab1').click()
    cy.get('[data-e2e=tab1]').should('be.visible')
  })

  it('Expand works', () => {
    cy.visit('/layout')
    cy.contains('Expand').should('exist')
  })

  it('Popover opens, survives a rerun, and closes on an outside click', () => {
    cy.visit('/layout')
    // The trigger carries the component's derived id, the way every other
    // component carries its Conf.ID.
    const trigger = () => cy.get('[id="popover_component_Advanced options"]')
    const inner = () => cy.get('.toolgui-popover-dropdown')
      .contains('button', 'Reset options')

    // The dropdown holds its contents from the start, hidden until the
    // trigger is clicked. Whether it is open is the trigger's aria-expanded:
    // the dropdown also hides itself whenever the trigger scrolls off the
    // screen, which is not the same thing as being closed.
    trigger().should('have.attr', 'aria-expanded', 'false')
    inner().should('not.be.visible')

    trigger().click()
    trigger().should('have.attr', 'aria-expanded', 'true')
    inner().should('be.visible')

    // A widget inside the popover reruns the page. The client owns whether
    // the popover is open, so the new props leave it open. Scrolled to the
    // middle rather than the top, which would push the trigger off the screen
    // and take the dropdown with it.
    inner().click({ scrollBehavior: 'center' })
    cy.get('#text_component_popover_reset').should('exist')
    trigger().should('have.attr', 'aria-expanded', 'true')
    inner().should('be.visible')

    // Anything outside the dropdown closes it.
    cy.get('.toolgui-box').contains('A box!').click()
    trigger().should('have.attr', 'aria-expanded', 'false')
  })

  // The slot is written three times in one run, and what it holds at the end
  // is the only thing on the screen. Everything is scoped to the component
  // column: the column beside it shows the source, which names the same text.
  it('Empty writes over its contents instead of stacking them', () => {
    cy.visit('/layout')
    const shown = () => cy.get('#column_component_show_empty_0')

    shown().contains('No query yet.').should('exist')
    shown().contains('Run a slow query').click()
    shown().contains('Querying…').should('exist')

    shown().contains('orders', { timeout: 15000 }).should('exist')
    shown().contains('Querying…').should('not.exist')
    shown().contains('No query yet.').should('not.exist')
  })
})
