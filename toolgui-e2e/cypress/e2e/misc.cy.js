describe('Misc', () => {
  // The message is shown in an alert, not dropped into the page as text. What
  // the page function returned is the page talking to its user, so it arrives
  // whole rather than masked.
  it('Error handling', () => {
    cy.visit('/misc')
    cy.contains('Show error').click()
    cy.get('[role=alert]').contains('new error').should('exist')
  })

  // A panic is the framework's failure, not the page's: the browser is told
  // the kind of error and an id to look the server log line up by, never what
  // the app panicked with.
  it('Panic handling', () => {
    cy.visit('/misc')
    cy.contains('Show panic').click()
    cy.get('[role=alert]').contains('internal error').should('exist')
    cy.get('[role=alert]').contains(/error id: \w+/).should('exist')
    cy.get('[role=alert]').contains('show panic').should('not.exist')
  })

  it('HTML component', () => {
    cy.visit('/misc')
    cy.get('b').contains('Hello world gen by html component').should('exist')
  })

  // Scoped to the component column: the column beside it shows the source,
  // which names the same text.
  it('Spinner comes down when the work is done', () => {
    cy.visit('/misc')
    const shown = () => cy.get('#column_component_show_spinner_0')

    shown().contains('Spin for three seconds').click()
    shown().contains('Working…').should('exist')

    shown().contains('Done!', { timeout: 15000 }).should('exist')
    shown().contains('Working…').should('not.exist')
  })

  // The bar's value lives in state, so each click has to survive the rerun.
  // The label is scoped to the component column, or the source beside it
  // matches too; the bar itself carries the value as aria-valuenow.
  it('Progress bar moves ten points per click', () => {
    cy.visit('/misc')
    const shown = () => cy.get('#column_component_show_progress_bar_0')
    const bar = () =>
      cy.get('#progress_bar_component_misc_progress [role=progressbar]')

    bar().should('have.attr', 'aria-valuenow', '30')
    shown().contains('progress_bar: 30%').should('exist')

    shown().contains('+10%').click()
    bar().should('have.attr', 'aria-valuenow', '40')
    shown().contains('progress_bar: 40%').should('exist')

    shown().contains('+10%').click()
    bar().should('have.attr', 'aria-valuenow', '50')
    shown().contains('progress_bar: 50%').should('exist')
  })

  // A toast is something that happened, so a node that renders nothing is
  // what it leaves behind, and every run that reaches the call fires it again.
  it('Toasts stack, go by themselves, and fire again on the next run', () => {
    cy.visit('/misc')
    const shown = () => cy.get('#column_component_show_toast_0')
    const toast = (text, opts) =>
      cy.contains('.mantine-Notification-root', text, opts)

    shown().contains('Save').click()

    // Two calls in one run, two toasts, both on screen at once.
    toast('Saved to disk').should('be.visible')
    toast('Two rows changed').should('be.visible')

    // Nothing landed where the page function wrote them.
    shown().should('not.contain', 'Saved to disk')

    // The first took the default duration and the second was given ten
    // seconds, so the first goes while the second is still up.
    toast('Saved to disk', { timeout: 10000 }).should('not.exist')
    toast('Two rows changed').should('be.visible')

    // Same call, same place, second run. The node is the one that is already
    // there, so only the run serial says this is a second toast.
    shown().contains('Save').click()
    toast('Saved to disk').should('be.visible')
  })

  // Pausing on hover is what gives someone time to read a toast that is about
  // to go.
  it('A hovered toast stays up past its duration', () => {
    cy.visit('/misc')
    const toast = (text, opts) =>
      cy.contains('.mantine-Notification-root', text, opts)

    cy.get('#column_component_show_toast_0').contains('Save').click()
    toast('Saved to disk').should('be.visible').trigger('mouseover')

    // Well past the four seconds it would otherwise have had.
    cy.wait(8000)
    toast('Saved to disk').should('be.visible')

    // And it goes once the pointer leaves.
    toast('Saved to disk').trigger('mouseout')
    toast('Saved to disk', { timeout: 10000 }).should('not.exist')
  })

  it('Status collects its lines and closes as a success', () => {
    cy.visit('/misc')
    const shown = () => cy.get('#column_component_show_status_0')

    shown().contains('Import three files').click()
    shown().contains('Importing…').should('exist')

    shown().contains('Imported 3 files', { timeout: 20000 }).should('exist')
    shown().contains('three.csv').should('exist')
    shown().contains('Importing…').should('not.exist')
  })
})
