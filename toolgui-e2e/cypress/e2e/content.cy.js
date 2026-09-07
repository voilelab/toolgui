describe('Content', () => {
  it('Title works', () => {
    cy.visit('/content')
    cy.get('h1').contains('Title').should('exist')
  })

  it('Subtitle works', () => {
    cy.visit('/content')
    cy.get('h2').contains('Subtitle').should('exist')
  })

  it('Text works', () => {
    cy.visit('/content')
    cy.contains('Text').should('exist')
  })

  it('Image works', () => {
    cy.visit('/content')
    cy.get('img').should('have.attr', 'src', 'https://http.cat/100')
  })

  // Mantine's Divider is a div with the separator role, the accessible
  // equivalent of the hr this used to render.
  it('Divider works', () => {
    cy.visit('/content')
    cy.get('[role=separator]').should('exist')
  })

  it('Link works', () => {
    cy.visit('/content')
    cy.get('a').contains('Link').should('exist')
  })

  it('Latex works', () => {
    cy.visit('/content')
    cy.get('mi').contains('E').should('exist')
  })

  // The right column echoes the source, which holds the shortcode as
  // written, so both assertions scope to the left column, the rendered one.
  it('Emoji shortcodes expand in text', () => {
    cy.visit('/content')
    cy.get('#column_component_show_emoji_0')
      .contains('Shipped it \u{1F389}').should('exist')
  })

  it('Emoji shortcodes stay literal in markdown code', () => {
    cy.visit('/content')
    cy.get('#column_component_show_emoji_0')
      .find('code').should('have.length', 1)
      .and('have.text', ':tada:')
  })

  it('Identical components both render', () => {
    cy.visit('/content')
    cy.get('#column_component_show_duplicate')
      .contains('written twice').should('exist')
    cy.get('#column_component_show_duplicate')
      .find('div').filter((i, el) => el.textContent.trim() === 'written twice')
      .should('have.length', 2)
  })
})
