describe('App', () => {
  it('The tab holds the page title and the app title', () => {
    cy.visit('/index')
    cy.title().should('eq', 'Index - ToolGUI Demo')

    cy.visit('/layout')
    cy.title().should('eq', 'Layout - ToolGUI Demo')
  })

  it('The manifest is the one the app configured', () => {
    cy.request('/manifest.json').then((resp) => {
      expect(resp.headers['content-type']).to.contain('application/manifest+json')

      // Cypress only parses application/json for us.
      const manifest = typeof resp.body === 'string' ?
        JSON.parse(resp.body) : resp.body

      expect(manifest.name).to.eq('ToolGUI Demo')
      expect(manifest.short_name).to.eq('ToolGUI')
      expect(manifest.display).to.eq('standalone')
    })
  })
})
