const { defineConfig } = require("cypress");

// The browser build of the demo, served as a static site. It is a separate run
// from the web one: the pages are the same, but the transport underneath is
// the wasm bridge rather than HTTP, and what is worth testing twice is the
// half that differs.
module.exports = defineConfig({
  e2e: {
    baseUrl: 'http://127.0.0.1:3000',
    specPattern: 'cypress/e2e-wasm/**/*.cy.js',
    // Booting the wasm binary and streaming a large upload both take longer
    // than a request to a server next door.
    defaultCommandTimeout: 120000,
    setupNodeEvents(on, config) {
      // implement node event listeners here
    },
  },
});
