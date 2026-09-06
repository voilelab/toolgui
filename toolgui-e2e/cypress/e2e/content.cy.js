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

  it('Divider works', () => {
    cy.visit('/content')
    cy.get('hr').should('exist')
  })

  it('Link works', () => {
    cy.visit('/content')
    cy.get('a').contains('Link').should('exist')
  })

  it('Latex works', () => {
    cy.visit('/content')
    cy.get('mi').contains('E').should('exist')
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
