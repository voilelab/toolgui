describe('Misc', () => {
  // The message is shown in an alert, not dropped into the page as text.
  it('Error handling', () => {
    cy.visit('/misc')
    cy.contains('Show error').click()
    cy.get('[role=alert]').contains('new error').should('exist')
  })

  it('Panic handling', () => {
    cy.visit('/misc')
    cy.contains('Show panic').click()
    cy.get('[role=alert]').contains('show panic').should('exist')
  })

  it('Html component', () => {
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
