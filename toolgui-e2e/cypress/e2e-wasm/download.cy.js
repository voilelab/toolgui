const path = require('path')

// The browser build has no server to fetch a download from: the bytes are in
// the origin private file system, the worker reads the file Go names with
// getFile, and the blob that comes back is what the page saves. Nothing about
// the content is in the pack, and no copy of it is ever a string in the tab --
// which is the same thing the upload spec next door checks on the way in.

const downloadsFolder = Cypress.config('downloadsFolder')

// What the demo's DownloadFile offers: a megabyte of a pattern, so a byte
// anywhere in the file is known from its offset alone.
const patternSize = 1024 * 1024

function pattern() {
  const bs = Cypress.Buffer.alloc(patternSize)
  for (let i = 0; i < bs.length; i++) {
    bs[i] = i % 251
  }

  return bs
}

describe('Wasm download', () => {
  it('saves every byte the page offered', () => {
    // A static host cannot route paths, so the browser build puts the page in
    // the hash.
    cy.visit('/#/input')

    // Once the binary has booted and the page has drawn itself.
    cy.get('button').contains('Save a megabyte').should('exist').click()

    // The page hears about the click once the bytes are in hand, so this is
    // the wait for the read as well.
    cy.contains('Megabyte saved!').should('exist')

    cy.readFile(path.join(downloadsFolder, 'pattern.bin'), null).then((got) => {
      expect(got.length).to.eq(patternSize)
      expect(Cypress.Buffer.compare(got, pattern())).to.eq(0)
    })
  })
})
