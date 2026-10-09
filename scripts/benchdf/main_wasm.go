//go:build js && wasm

package main

import "github.com/voilelab/toolgui/toolgui/tgwasm"

func main() {
	app := newApp()
	app.SetHashPageNameMode(true)
	tgwasm.NewExecutor(app).Run()
}
