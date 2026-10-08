//go:build !(js && wasm)

package main

import (
	"flag"
	"log"

	"github.com/voilelab/toolgui/toolgui/tgexec"
)

func main() {
	addr := flag.String("addr", "127.0.0.1:3100", "listen address")
	flag.Parse()

	e := tgexec.NewWebExecutor(newApp())
	log.Println("listening on", *addr)
	log.Fatal(e.StartService(*addr))
}
