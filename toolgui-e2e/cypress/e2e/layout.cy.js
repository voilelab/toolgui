// A Mantine Modal and a Mantine Popover dropdown both carry role=dialog, and
// this page has one of each. aria-modal is what tells them apart: only the
// Modal sets it.
const DIALOG = '[role=dialog][aria-modal=true]'

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

  it('Toolbar lines its items up in one row', () => {
    cy.visit('/toolbar')

    // One row means one top edge. Written straight into the page each of
    // these would take a row to itself, which is the whole point of the
    // component.
    cy.get('#button_component_toolbar_run').then($run => {
      const top = $run[0].getBoundingClientRect().top

      cy.get('#button_component_toolbar_stop').should($stop => {
        expect($stop[0].getBoundingClientRect().top).to.be.closeTo(top, 2)
      })
    })
  })

  it('A sticky toolbar stays at the top of a scrolled page', () => {
    cy.visit('/toolbar')

    const bar = () => cy.get('#toolbar_component_sticky_toolbar')
    bar().should('have.css', 'position', 'sticky')

    // The rows under it are what the page scrolls past.
    cy.contains('row-39').scrollIntoView()

    bar().should($bar => {
      const rect = $bar[0].getBoundingClientRect()
      expect(rect.top).to.be.closeTo(0, 2)
      expect(rect.height).to.be.greaterThan(0)
    })

    // Opaque, or the rows passing under would read through the row.
    bar().should('not.have.css', 'background-color', 'rgba(0, 0, 0, 0)')
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

  // Mantine hangs a popover's ESC off the dropdown, so it only fires while
  // focus is already inside the panel -- not after the trigger opened it.
  it('Popover closes on ESC from its trigger', () => {
    cy.visit('/layout')
    const trigger = () => cy.get('[id="popover_component_Advanced options"]')

    trigger().click()
    trigger().should('have.attr', 'aria-expanded', 'true')

    // No click into the panel first: focus is still on the button.
    cy.get('body').type('{esc}')
    trigger().should('have.attr', 'aria-expanded', 'false')

    trigger().click()
    trigger().should('have.attr', 'aria-expanded', 'true')
  })

  // ESC goes to whichever overlay opened last, whatever kind it is.
  it('A popover inside a dialog takes the ESC before the dialog does', () => {
    cy.visit('/layout')
    const trigger = () => cy.get('[id="popover_component_Delete options"]')

    cy.get('#column_component_show_dialog_0').contains('Delete').click()
    cy.get(DIALOG).should('have.length', 1)

    trigger().click()
    trigger().should('have.attr', 'aria-expanded', 'true')

    cy.get('body').type('{esc}')
    trigger().should('have.attr', 'aria-expanded', 'false')
    cy.get(DIALOG).should('have.length', 1)

    // With the popover gone the dialog is topmost again.
    cy.get('body').type('{esc}')
    cy.get(DIALOG).should('not.exist')
  })

  // Text assertions here are scoped to the dialog itself: the column beside
  // the component shows the source, which names the same strings.
  it('Dialog draws nothing until something opens it', () => {
    cy.visit('/layout')
    cy.get(DIALOG).should('not.exist')

    cy.get('#column_component_show_dialog_0').contains('Delete').click()
    cy.get(DIALOG).contains('Delete the 3 selected rows?')
      .should('be.visible')
  })

  it('Dialog closes from a button inside it', () => {
    cy.visit('/layout')
    cy.get('#column_component_show_dialog_0').contains('Delete').click()
    cy.get(DIALOG).contains('Yes, delete').click()

    // Closed by the run that handled the click, and the run after it does not
    // write the body again.
    cy.get(DIALOG).should('not.exist')
    cy.contains('button', 'Rerun').click()
    cy.get(DIALOG).should('not.exist')
  })

  it('Dialog closed with ESC stays closed on the next run', () => {
    cy.visit('/layout')
    cy.get('#column_component_show_dialog_0').contains('Delete').click()
    cy.get(DIALOG).should('be.visible')

    cy.get('body').type('{esc}')
    cy.get(DIALOG).should('not.exist')

    // Whether it is open is kept by the page, so this is also the check that
    // the dismissal reached the server rather than only the client.
    cy.contains('button', 'Rerun').click()
    cy.get(DIALOG).should('not.exist')
  })

  it('A dialog opened from another one stacks over it', () => {
    cy.visit('/layout')
    cy.get('#column_component_show_dialog_0').contains('Delete').click()
    cy.get(DIALOG).contains('What does this do?').click()

    cy.get(DIALOG).should('have.length', 2)

    // Visible is not the same as on top, so ask the document what is at the
    // middle of the second dialog. It is the one the page writes FIRST, so
    // this fails unless being on top follows the order they opened.
    cy.contains(DIALOG, 'The rows are removed for good.').then(($d) => {
      const box = $d[0].getBoundingClientRect()
      cy.document().then((doc) => {
        const at = doc.elementFromPoint(
          box.x + box.width / 2, box.y + box.height / 2)
        expect(at.closest(DIALOG)).to.contain.text('removed for good')
      })
    })

    cy.get(DIALOG).contains('Got it').click()
    cy.get(DIALOG).should('have.length', 1)
    cy.get(DIALOG).contains('Delete the 3 selected rows?')
      .should('be.visible')
  })

  // Mantine gives every Modal its own window key handler with no notion of a
  // stack, so leaving closeOnEscape on closed the whole stack at once. The
  // dialog this closes is the one drawn on top, which is why the same stack
  // drives the z-index.
  it('ESC closes the top dialog only', () => {
    cy.visit('/layout')
    cy.get('#column_component_show_dialog_0').contains('Delete').click()
    cy.get(DIALOG).contains('What does this do?').click()
    cy.get(DIALOG).should('have.length', 2)

    cy.get('body').type('{esc}')
    cy.get(DIALOG).should('have.length', 1)

    // The ESC sent an input event, so the page reran and re-sent both dialogs
    // with its own idea of open. The parent still being here after that round
    // trip says the page did not close it either.
    cy.get(DIALOG).contains('Delete the 3 selected rows?')
      .should('be.visible')

    cy.get('body').type('{esc}')
    cy.get(DIALOG).should('not.exist')
  })

  it('Dialog written into the sidebar still covers the window', () => {
    cy.visit('/layout')
    cy.contains('button', 'Open the sidebar dialog').click()

    cy.get(DIALOG).contains('Declared in the sidebar')
      .should('be.visible')

    // A portal: drawn at the top of the document, not inside the side column
    // the page wrote it into.
    cy.get(`#container_component_container_sidebar ${DIALOG}`)
      .should('not.exist')

    // Dismissible is off on this one: no X, and ESC leaves it alone.
    cy.get(DIALOG).find('[aria-label="Close dialog"]').should('not.exist')
    cy.get('body').type('{esc}')
    cy.get(DIALOG).should('be.visible')

    cy.contains('button', 'Close the sidebar dialog').click()
    cy.get(DIALOG).should('not.exist')
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
