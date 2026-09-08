//go:build !(js && wasm)

package main

import (
	"log"

	"github.com/voilelab/toolgui/toolgui/tgexec"
)

func main() {
	app := newApp()

	if err := addPluginDemo(app); err != nil {
		log.Fatal(err)
	}

	e := tgexec.NewWebExecutor(app)

	// What a browser reads when the demo is installed to a home screen. It is
	// a WebExecutor setting, so the browser build does without it.
	e.SetManifest(&tgexec.Manifest{
		Name:            "ToolGUI Demo",
		ShortName:       "ToolGUI",
		Description:     "A demo of the components ToolGUI provides.",
		StartURL:        ".",
		Display:         "standalone",
		ThemeColor:      "#000000",
		BackgroundColor: "#ffffff",
	})

	log.Println("Starting service...")

	err := e.StartService("127.0.0.1:3000")
	if err != nil {
		log.Println(err)
	}
}
