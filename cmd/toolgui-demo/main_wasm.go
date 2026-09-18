//go:build js && wasm

package main

import "github.com/voilelab/toolgui/toolgui/tgwasm"

// Embedded, this build is in an iframe under a book page that already carries
// the example's code, so an example draws the component alone rather than
// splitting the frame's width with a copy of it. On its own -- the demo app,
// published beside the book -- nothing else carries the code, so it does.
var withCode = !tgwasm.Embedded()

// The browser build, published to GitHub Pages beside the book.
func main() {
	app := newApp()

	// Pages cannot route paths under a project site, so pages live in the
	// hash.
	app.SetHashPageNameMode(true)

	tgwasm.NewExecutor(app).Run()
}
