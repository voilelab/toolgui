//go:build js && wasm

// Package tgwasm runs a ToolGUI app in the browser, compiled to WebAssembly.
//
// It's the browser-only counterpart of tgexec.WebExecutor: same app, same
// pages, same frontend components, with calls on a JavaScript object in place
// of HTTP and websockets. There is no server, so the app ships as static
// files.
//
//	//go:build js && wasm
//
//	app := tgframe.NewApp()
//	app.AddPage("index", "Index", Main)
//	tgwasm.NewExecutor(app).Run()
package tgwasm

import "github.com/voilelab/toolgui/toolgui/tgframe"

// Executor is a browser ui executor for ToolGUI.
type Executor struct {
	app *tgframe.App
}

// NewExecutor return an Executor serving app in the page that loaded it.
func NewExecutor(app *tgframe.App) *Executor {
	return &Executor{app: app}
}

// Run install the bridge and block forever, so the wasm instance stays alive
// to answer the page. It never returns.
func (e *Executor) Run() {
	newBridge(e.app).install()

	select {}
}
