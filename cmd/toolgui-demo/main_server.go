//go:build !(js && wasm)

package main

import (
	"log"

	"github.com/voilelab/toolgui/toolgui/tgexec"
)

func main() {
	e := tgexec.NewWebExecutor(newApp())
	log.Println("Starting service...")

	err := e.StartService(":3000")
	if err != nil {
		log.Println(err)
	}
}
