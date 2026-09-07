//go:build js && wasm

package main

import "github.com/voilelab/toolgui/toolgui/tgwasm"

// The browser build, published to GitHub Pages beside the book.
func main() {
	app := newApp()

	// Pages cannot route paths under a project site, so pages live in the
	// hash.
	app.SetHashPageNameMode(true)

	tgwasm.NewExecutor(app).Run()
}
