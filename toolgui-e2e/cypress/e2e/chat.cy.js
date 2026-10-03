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
})
