describe('Chat', () => {
  it('Messages are drawn by role, with what they hold', () => {
    cy.visit('/chat')
    const shown = () => cy.get('#column_component_show_chat_message_0')

    shown().find('.toolgui-chat-message').should('have.length', 3)
    shown().find('.toolgui-chat-message-user')
      .should('have.attr', 'aria-label', 'user message')
      .contains('What can ChatMessage hold?')
    shown().find('.toolgui-chat-message-assistant').contains('1,024')
    shown().find('.toolgui-chat-message-other').contains('🔧')
  })

  // The reply grows while the run is still going, and ends whole.
  it('A reply streams into its message', () => {
    cy.visit('/chat')
    const shown = () => cy.get('#column_component_show_chat_message_stream_0')

    shown().contains('Ask').click()
    shown().find('.toolgui-chat-message-assistant')
      .contains('WriteStream writes').should('exist')
    shown().find('.toolgui-chat-message-assistant')
      .contains('once the stream ends.', { timeout: 10000 }).should('exist')
  })

  // Enter sends and empties the box; the reply streams in, and the turns are
  // kept for the next send.
  it('A sent message is answered and kept', () => {
    cy.visit('/chat')
    const shown = () => cy.get('#column_component_show_chat_input_0')
    const box = () => shown().find('textarea').first()

    shown().find('button[aria-label=Send]').should('be.disabled')

    box().type('hello{enter}')
    box().should('have.value', '')
    shown().find('.toolgui-chat-message-assistant')
      .contains('You said: hello', { timeout: 10000 }).should('exist')

    box().type('again')
    shown().find('button[aria-label=Send]').click()
    shown().find('.toolgui-chat-message-assistant')
      .contains('You said: again', { timeout: 10000 }).should('exist')
    shown().find('.toolgui-chat-message').should('have.length', 4)
  })
})
