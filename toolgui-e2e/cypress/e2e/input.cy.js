const path = require('path')

describe('Input', () => {
  it('Textarea input', () => {
    cy.visit('/input')
    cy.get('textarea').type('testarea: 1')
    cy.get('textarea').blur()
    cy.get('textarea').type('2')
    cy.get('textarea').type('3')
    cy.get('textarea').blur()
    cy.contains('Value: testarea: 123')
  })

  it('Textbox input', () => {
    cy.visit('/input')
    cy.get('input[id=textbox_component_Textbox]').type('abc')
    cy.get('input[id=textbox_component_Textbox]').blur()
    cy.get('input[id=textbox_component_Textbox]').type('abc')
    cy.get('input[id=textbox_component_Textbox]').blur()
    cy.contains('Value: abcabc')
  })

  it('Fileupload input', () => {
    cy.visit('/input')
    // accept only filters the file picker, so it is the whole of the limit.
    cy.get('input[type=file]').should('have.attr', 'accept', '.jpg,.png')

    // Mantine keeps the file input hidden behind its own control, so the
    // file is handed to the input itself.
    cy.get('input[type=file]').selectFile('cypress/fixtures/example.png', {
      force: true,
    })
    cy.contains('Fileupload filename: example.png').should('exist')
    // The bytes reach Go, not just the name.
    cy.contains(/Fileupload bytes length: [1-9]\d*/).should('exist')
  })

  it('Checkbox', () => {
    cy.visit('/input')
    cy.get('input[type=checkbox]').click()
    cy.contains('Value: true').should('exist')
  })

  it('Button click', () => {
    cy.visit('/input')
    cy.contains('button').click()
    cy.contains('Value: true').should('exist')
    cy.contains('Rerun').click()
    cy.get('div[id=column_component_show_button]').contains('Value: false').should('exist')
  })

  it('Select', () => {
    cy.visit('/input')
    // Mantine's Select is a combobox over a listbox, not a native <select>:
    // open the dropdown, then click the option.
    //
    // The option clicks turn scrolling off. Cypress scrolls what it is about
    // to click to the top of the viewport, and the option sits in a portal
    // below the input — scrolling it up carries the input off screen, at
    // which point Mantine hides the dropdown as detached from it. The
    // dropdown opens in view already, so there is nothing to scroll to.
    const noScroll = { scrollBehavior: false }

    cy.get('input[id=select_component_Select]').click()
    cy.get('[role=option]').contains('Value1').click(noScroll)
    cy.contains('Value: Value1').should('exist')

    cy.get('input[id=select_component_Select]').click()
    cy.get('[role=option]').contains('Value2').click(noScroll)
    cy.contains('Value: Value2').should('exist')
  })

  it('Radio', () => {
    cy.visit('/input')
    cy.contains('Value3').click()
    cy.contains('Value: Value3').should('exist')

    cy.contains('Value4').click()
    cy.contains('Value: Value4').should('exist')
  })

  it('Datepicker', () => {
    cy.visit('/input')
    // Mantine's DateInput is a text input that parses what is typed, so there
    // is no type=date to select on and no segments to overwrite — the second
    // date has to clear the first.
    const date = 'input[id=datepicker_component_Datepicker]'

    cy.get(date).type('2000-01-01')
    cy.get(date).blur()
    cy.contains('2000-01-01').should('exist')

    cy.get(date).clear()
    cy.get(date).type('2002-02-02')
    cy.get(date).blur()
    cy.contains('2002-02-02').should('exist')
  })

  it('Timepicker', () => {
    cy.visit('/input')
    cy.get('input[type=time]').type('20:34')
    cy.get('input[type=time]').blur()
    cy.contains('20:34').should('exist')

    cy.get('input[type=time]').type('11:34')
    cy.get('input[type=time]').blur()
    cy.contains('11:34').should('exist')
  })

  it('Datetimepicker', () => {
    cy.visit('/input')

    // Mantine's DateTimePicker is a calendar popover, not a typeable
    // input[type=datetime-local]. Everything inside the popover is clicked
    // with scrolling off, for the reason given in the Select case above.
    const noScroll = { scrollBehavior: false }

    // The calendar opens on the current month, so the dates come from this
    // month rather than a fixed year.
    const dayOfThisMonth = (day) => {
      const date = new Date()
      date.setDate(day)
      return date
    }

    // How Mantine labels a day cell, e.g. '1 September 2026'.
    const dayLabel = (date) => date.toLocaleDateString('en-GB', {
      day: 'numeric', month: 'long', year: 'numeric',
    })

    // What the demo prints back, e.g. '2026-09-01 20:34'.
    const expected = (date, time) => {
      const pad = (n) => String(n).padStart(2, '0')
      return `${date.getFullYear()}-${pad(date.getMonth() + 1)}` +
        `-${pad(date.getDate())} ${time}`
    }

    const openAndPickDay = (date) => {
      cy.get('button[id=datepicker_component_Datetimepicker]').click()
      cy.get(`button[aria-label="${dayLabel(date)}"]`).click(noScroll)
    }

    const submit = () => {
      cy.get('.mantine-DateTimePicker-submitButton').click(noScroll)
    }

    // The time goes in once, while the hour and minute fields are still
    // empty: they hold a digit buffer that typing over a value it was given
    // rather than typed does not overwrite predictably.
    const first = dayOfThisMonth(1)
    openAndPickDay(first)
    cy.get('[role=spinbutton]').eq(0).type('20', noScroll)
    cy.get('[role=spinbutton]').eq(1).type('34', noScroll)
    submit()
    cy.contains(expected(first, '20:34')).should('exist')

    // A second value round-trips too. Only the date moves; the time it keeps
    // is what makes that visible without retyping it.
    const second = dayOfThisMonth(2)
    openAndPickDay(second)
    submit()
    cy.contains(expected(second, '20:34')).should('exist')
  })

  it('Number', () => {
    cy.visit('/input')
    cy.get('input[id=number_component_Number]').type('{backspace}2')
    cy.get('input[id=number_component_Number]').blur()
    cy.contains('Value: 12').should('exist')

    // 123 is over the max of 20, so the input is invalid and never sent.
    // Mantine's NumberInput is a text input that reports the range itself,
    // so the mark to check is the one it puts on the field.
    cy.get('input[id=number_component_Number]').type('{backspace}23')
    cy.get('input[id=number_component_Number]').blur()
    cy.get('input[id=number_component_Number]')
      .should('have.value', '123')
      .and('have.attr', 'aria-invalid', 'true')
    cy.contains('Value out of range').should('exist')
    cy.contains('Value: 123').should('not.exist')
  })

  it('Form', () => {
    cy.visit('/input')
    cy.get('input[id=number_component_a]').type('12')
    cy.get('input[id=number_component_b]').type('12')
    cy.contains('Submit').click()
    cy.contains('int(a) + int(b) = 24').should('exist')

    cy.get('input[id=number_component_b]').type('{backspace}3')
    cy.contains('int(a) + int(b) = 24').should('exist')
    cy.contains('Submit').click()
    cy.contains('int(a) + int(b) = 25').should('exist')
  })

  const downloadsFolder = Cypress.config('downloadsFolder');

  it('Download Button works', () => {
    cy.visit('/input')
    cy.get('button').contains('Download').click()

    cy.readFile(path.join(downloadsFolder, '123.txt')).should('equal', '123')
    cy.contains('Downloaded!').should('exist')
  })
})
