// The browser build keeps uploads in the origin private file system and gets
// them there by streaming: the worker pipes the picked file straight into a
// file and hands Go a handle on what it wrote. Nothing holds the file.
//
// What that buys is this spec. The transport used to make the whole upload one
// base64 string for Go to decode, which meant several copies of it in the tab
// at once, two of them a third larger again -- so a file of a few hundred
// megabytes took the tab down instead of arriving.

// uploadSize is comfortably past what the old transport could carry, and past
// what the tab would hold copies of.
const uploadSize = 200 * 1024 * 1024

// chunkSize is how much of the file is built at a time. The parts are the same
// array over and over, and a Blob keeps its bytes of its own accord, so the
// file is never built in memory here either.
const chunkSize = 1024 * 1024

// pick hands a file of the given size to the upload input, the way selectFile
// does. Building it in the page rather than reading a fixture is the point: a
// fixture this size would be read into memory, which is what the transport
// stopped doing.
function pick(size, name) {
  cy.get('input[type=file]').then(($input) => {
    const input = $input[0]
    const win = input.ownerDocument.defaultView

    const chunk = new win.Uint8Array(Math.min(size, chunkSize))
    for (let i = 0; i < chunk.length; i++) {
      chunk[i] = i % 251
    }

    const parts = new Array(Math.max(1, size / chunkSize)).fill(chunk)
    const file = new win.File(parts, name, { type: 'application/octet-stream' })

    const handover = new win.DataTransfer()
    handover.items.add(file)
    input.files = handover.files

    input.dispatchEvent(new win.Event('input', { bubbles: true }))
    input.dispatchEvent(new win.Event('change', { bubbles: true }))
  })
}

describe('Wasm upload', () => {
  it('takes a file far larger than the tab could hold copies of', () => {
    // A static host cannot route paths, so the browser build puts the page in
    // the hash.
    cy.visit('/#/input')

    // Mantine keeps the file input hidden behind its own control, so it is
    // there rather than visible, and it appears once the binary has booted.
    cy.get('input[type=file]').should('exist')

    pick(uploadSize, 'big.bin')

    cy.contains('FileUpload filename: big.bin').should('exist')

    // Every byte of it, so a stream that stopped early or wrote over itself is
    // not mistaken for one that arrived.
    cy.contains(`FileUpload bytes length: ${uploadSize}`).should('exist')

    // A second pick right after: the tab is still there to use, and the file
    // the component held is replaced rather than added to.
    pick(16, 'small.bin')

    cy.contains('FileUpload filename: small.bin').should('exist')
    cy.contains('FileUpload bytes length: 16').should('exist')
  })
})
