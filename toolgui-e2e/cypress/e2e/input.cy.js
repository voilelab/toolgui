const path = require('path')

describe('Input', () => {
  // Mantine follows the pointer from the track container rather than from the
  // component's root, which is where the id sits.
  const TRACK = '.mantine-Slider-trackContainer'

  // Steps a slider by one step. The keys are handled on the component's own
  // root, which is where the id sits, so the event goes straight there rather
  // than through the thumb — the thumb is a div, which cy.type() will not
  // take.
  const arrow = (selector, key) =>
    cy.get(selector).trigger('keydown', { key })

  // Presses at `from`, moves to `to`, waits for `beforeRelease` and lets go,
  // as fractions of the element's width. The moves are triggered on the
  // element itself and reach Mantine's document listeners by bubbling.
  //
  // beforeRelease is not optional. Mantine samples the pointer in a
  // requestAnimationFrame and reads the last sample when the button comes up,
  // so a release in the same tick as the move reports where the drag started
  // rather than where it ended. Assert something there that only holds once
  // that frame has landed -- which is also where a value that is reported
  // only on release can be caught not having been reported yet.
  const drag = (selector, from, to, beforeRelease, fy = 0.5) => {
    cy.get(selector).then(($el) => {
      const rect = $el[0].getBoundingClientRect()
      const at = (fx) => ({
        clientX: rect.left + rect.width * fx,
        clientY: rect.top + rect.height * fy,
        force: true,
      })

      cy.wrap($el)
        .trigger('mousedown', { button: 0, ...at(from) })
        .trigger('mousemove', at(to))

      beforeRelease()

      cy.wrap($el).trigger('mouseup', at(to))
    })
  }

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
    // By id rather than by type: the default-on checkbox below and the
    // Mantine Switches further down the page are all input[type=checkbox].
    cy.get('input[id=checkbox_component_Checkbox]').click()
    cy.get('div[id=column_component_show_checkbox]')
      .contains('Value: true').should('exist')
  })

  // A Default: true checkbox has to read true in Go on the first load, and
  // still be something the user can turn off for good.
  it('Checkbox default', () => {
    cy.visit('/input')
    const box = 'input[id="checkbox_component_Checkbox default on"]'

    cy.get(box).should('be.checked')
    cy.contains('Default: true').should('exist')

    cy.get(box).click()
    cy.contains('Default: false').should('exist')
    cy.get(box).should('not.be.checked')

    cy.contains('Rerun').click()
    cy.contains('Default: false').should('exist')
    cy.get(box).should('not.be.checked')
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

  it('Multiselect', () => {
    cy.visit('/input')
    // Same reason as the Select case: the options sit in a portal below the
    // input, and scrolling one to the top would take the input off screen.
    const noScroll = { scrollBehavior: false }
    // cy.contains(selector, text) yields the option itself; chaining contains
    // off a get would descend to the bare span holding the label, which
    // carries none of the option's attributes.
    const option = (label) => cy.contains('[role=option]', label)
    // Read off the demo's own line rather than the page: the code column
    // beside it prints the same words as source.
    const result = () => cy.get('#text_component_multiselect_result')

    // The dropdown stays open across picks, so it is opened once.
    cy.get('input[id=multiselect_component_Multiselect]').click()

    option('Alpha').click(noScroll)
    result().should('have.text', 'Values: Alpha')

    // A second pick joins the first rather than replacing it, and the result
    // is in item order whatever order they were picked in.
    option('Gamma').click(noScroll)
    result().should('have.text', 'Values: Alpha, Gamma')

    // MaxSelections is 2, so the item left over is disabled in the dropdown
    // instead of being taken and then refused.
    option('Beta').should('have.attr', 'data-combobox-disabled')

    // Deselecting an item frees the cap again.
    option('Alpha').click(noScroll)
    result().should('have.text', 'Values: Gamma')
    option('Beta').should('not.have.attr', 'data-combobox-disabled')

    // Deselecting the last one is an empty selection, not a fall back to the
    // first item.
    option('Gamma').click(noScroll)
    result().invoke('text').should('match', /^Values:\s*$/)
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

    // What the picker shows, and what the demo prints back after it,
    // e.g. '2026-09-01 20:34'.
    const shown = (date, time) => {
      const pad = (n) => String(n).padStart(2, '0')
      return `${date.getFullYear()}-${pad(date.getMonth() + 1)}` +
        `-${pad(date.getDate())} ${time}`
    }

    const picker = () => cy.get('button[id=datepicker_component_Datetimepicker]')
    const hours = () => cy.get('[role=spinbutton]').eq(0)
    const minutes = () => cy.get('[role=spinbutton]').eq(1)

    // Picking a day fills in the date and leaves the time to the fields
    // below, so the button reads back where the picker stands. Waiting for it
    // keeps the rest off a picker that has not taken the pick: with no date
    // of its own, Mantine dates the time from today, and the value comes back
    // right time, wrong day.
    const openAndPickDay = (date, time) => {
      picker().click()
      cy.get(`button[aria-label="${dayLabel(date)}"]`).click(noScroll)
      picker().should('have.text', shown(date, time))
    }

    // The hour and minute fields hold a digit buffer, and picking a day fills
    // them with 00. Typing onto the end of that reads as '002', which Mantine
    // takes for a finished hour and answers by moving on to the minutes, so
    // the second digit lands in the wrong field and the hour stays 02.
    // Selecting first means each field sees only what is typed.
    const typeTime = (field, digits) => {
      field().type(`{selectall}${digits}`, noScroll)
      field().should('have.value', digits)
    }

    const submit = () => {
      cy.get('.mantine-DateTimePicker-submitButton').click(noScroll)
    }

    const first = dayOfThisMonth(1)
    openAndPickDay(first, '00:00')
    typeTime(hours, '20')
    typeTime(minutes, '34')
    submit()
    // 'Value: ' is the demo printing the value back from Go. Without it the
    // picker's own button carries the same text, and the case passes on the
    // browser alone.
    cy.contains(`Value: ${shown(first, '20:34')}`).should('exist')

    // A second value round-trips too. Only the date moves; the time it keeps
    // is what makes that visible without retyping it.
    const second = dayOfThisMonth(2)
    openAndPickDay(second, '20:34')
    submit()
    cy.contains(`Value: ${shown(second, '20:34')}`).should('exist')
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
    const noScroll = { scrollBehavior: false }

    cy.get('input[id=number_component_a]').type('12')
    cy.get('input[id=number_component_b]').type('12')

    // The multiselect holds its pick until Submit, like every other field in
    // a form, and reaches Go with the rest of them.
    cy.get('input[id=multiselect_component_ops]').click()
    cy.contains('[role=option]', 'sum').click(noScroll)
    // The dropdown stays open across picks, and would cover Submit.
    cy.get('input[id=multiselect_component_ops]').blur()
    cy.contains('int(a) + int(b) = 24').should('not.exist')

    cy.contains('Submit').click()
    cy.contains('int(a) + int(b) = 24').should('exist')

    cy.get('input[id=number_component_b]').type('{backspace}3')
    cy.contains('int(a) + int(b) = 24').should('exist')
    cy.contains('Submit').click()
    cy.contains('int(a) + int(b) = 25').should('exist')

    // A second pick is carried alongside the first.
    cy.get('input[id=multiselect_component_ops]').click()
    cy.contains('[role=option]', 'product').click(noScroll)
    cy.get('input[id=multiselect_component_ops]').blur()
    cy.contains('Submit').click()
    cy.contains('int(a) + int(b) = 25').should('exist')
    cy.contains('int(a) * int(b) = 156').should('exist')
  })

  it('Slider', () => {
    cy.visit('/input')

    // Mantine's Slider is a div, so the id is on its root. The track
    // container inside it is the part that follows the pointer, and the thumb
    // is what reports where the pointer has taken it.
    const root = 'div[id=slider_component_Slider]'
    const track = `${root} ${TRACK}`
    const thumb = `${root} [role=slider]`
    const result = () => cy.get('div[id=column_component_show_slider]')

    // 'Value: ' is the demo printing the value back from Go. Without it the
    // slider's own bubble carries the same number, and the case would pass on
    // the browser alone.
    result().contains('Value: 50').should('exist')

    // The step is 10, so an arrow key is one step.
    arrow(root, 'ArrowRight')
    result().contains('Value: 60').should('exist')

    arrow(root, 'ArrowLeft')
    arrow(root, 'ArrowLeft')
    result().contains('Value: 40').should('exist')

    // A drag reports on release, not on every step it crosses, which is what
    // keeps one drag to one rerun.
    drag(track, 0.4, 0.9, () => {
      // Mid-drag: the handle is at the end of the track, and nothing has been
      // sent, so Go is still reporting where the drag started.
      cy.get(thumb).should('have.attr', 'aria-valuenow', '90')
      result().contains('Value: 40').should('exist')
    })
    result().contains('Value: 90').should('exist')
  })

  it('SelectSlider', () => {
    cy.visit('/input')

    const root = 'div[id=select_slider_component_SelectSlider]'
    const result = () => cy.get('div[id=column_component_show_select_slider]')

    // Nothing picked yet, so the handle is on the first item.
    result().contains('Value: S').should('exist')

    arrow(root, 'ArrowRight')
    result().contains('Value: M').should('exist')

    arrow(root, 'ArrowRight')
    result().contains('Value: L').should('exist')

    // The last item is the end of the track: there is nothing past it to
    // slide onto, so the value stays where it is.
    arrow(root, 'ArrowRight')
    result().contains('Value: L').should('exist')

    arrow(root, 'ArrowLeft')
    arrow(root, 'ArrowLeft')
    result().contains('Value: S').should('exist')

    // And a drag lands on the mark it is let go over.
    drag(`${root} ${TRACK}`, 1, 1, () => {
      cy.get(`${root} [role=slider]`).should('have.attr', 'aria-valuenow', '2')
    })
    result().contains('Value: L').should('exist')
  })

  it('Toggle', () => {
    cy.visit('/input')

    // Mantine draws the switch itself and leaves the checkbox behind it at
    // zero opacity, so the click goes to the input the way the file one does.
    const toggle = 'input[id=toggle_component_Toggle]'
    const result = () => cy.get('div[id=column_component_show_toggle]')

    result().contains('Value: false').should('exist')

    // Both halves each time: the switch itself moves, and Go agrees. A
    // controlled Switch whose state does not move would pass the second
    // check and fail the first.
    cy.get(toggle).click({ force: true })
    cy.get(toggle).should('be.checked')
    result().contains('Value: true').should('exist')

    cy.get(toggle).click({ force: true })
    cy.get(toggle).should('not.be.checked')
    result().contains('Value: false').should('exist')
  })

  it('ColorPicker', () => {
    cy.visit('/input')

    const picker = 'input[id=color_picker_component_ColorPicker]'
    const result = () => cy.get('div[id=column_component_show_color_picker]')

    // Untouched, the value is the conf's default.
    result().contains('Value: #ff3860').should('exist')
    cy.get(picker).should('have.value', '#ff3860')

    // Typed in uppercase, read back in lower: Go promises '#rrggbb'.
    cy.get(picker).clear()
    cy.get(picker).type('#23D160')
    cy.get(picker).blur()
    result().contains('Value: #23d160').should('exist')

    // Picked off the saturation area rather than typed. Which color a point
    // in that area is, is the browser's business, so what is checked is that
    // Go ends up reporting the one the picker is left showing.
    cy.get(picker).click()
    cy.get(picker).invoke('val').then((typed) => {
      drag('.mantine-ColorInput-saturation', 0.9, 0.9, () => {
        // Mid-drag: the field has moved off the typed color, and nothing has
        // been sent, so Go is still reporting the typed one.
        cy.get(picker).should('not.have.value', typed)
        result().contains('Value: #23d160').should('exist')
      }, 0.9)
    })

    cy.get(picker).invoke('val').then((picked) => {
      result().contains(`Value: ${picked.toLowerCase()}`).should('exist')
    })
  })

  it('Form carries the new widgets', () => {
    cy.visit('/input')

    const result = () => cy.get('div[id=column_component_show_widget_form]')

    const thumb = 'div[id=slider_component_threshold] [role=slider]'
    const toggle = 'input[id=toggle_component_enabled]'

    result().contains('threshold = 0, enabled = false').should('exist')

    // Inside a form nothing reruns until Submit, so the values move on screen
    // while Go keeps reporting the ones it last received. The on-screen half
    // is the one worth asserting here: with no rerun to redraw them, a
    // controlled input that does not keep its own state sits there refusing
    // to move, and only the form case shows it.
    arrow('div[id=slider_component_threshold]', 'ArrowRight')
    cy.get(thumb).should('have.attr', 'aria-valuenow', '25')

    cy.get(toggle).click({ force: true })
    cy.get(toggle).should('be.checked')

    // It goes back and forth, not just on once.
    cy.get(toggle).click({ force: true })
    cy.get(toggle).should('not.be.checked')
    cy.get(toggle).click({ force: true })
    cy.get(toggle).should('be.checked')

    // And through all of that Go has heard nothing.
    result().contains('threshold = 0, enabled = false').should('exist')

    result().contains('Submit').click()
    result().contains('threshold = 25, enabled = true').should('exist')
  })

  const downloadsFolder = Cypress.config('downloadsFolder');

  it('Download Button works', () => {
    cy.visit('/input')
    cy.get('button').contains('Download').click()

    cy.readFile(path.join(downloadsFolder, '123.txt')).should('equal', '123')
    cy.contains('Downloaded!').should('exist')
  })
})
