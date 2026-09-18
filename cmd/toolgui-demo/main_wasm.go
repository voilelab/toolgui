//go:build js && wasm

package main

import "github.com/voilelab/toolgui/toolgui/tgwasm"

// The book embeds this build in an iframe, under a page that already carries
// the example's code, so an example here draws the component alone rather
// than splitting the frame's width with a copy of it.
const withCode = false

// The browser build, published to GitHub Pages beside the book.
func main() {
	app := newApp()

	// Pages cannot route paths under a project site, so pages live in the
	// hash.
	app.SetHashPageNameMode(true)

	tgwasm.NewExecutor(app).Run()
}
