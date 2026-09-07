// Command toolgui-wasm builds a ToolGUI app into a static site: your Go
// program compiled to WebAssembly, the frontend that drives it, and the
// wasm_exec.js of the toolchain that compiled it.
//
//	go get -tool github.com/voilelab/toolgui/cmd/toolgui-wasm
//	go tool toolgui-wasm build -o dist ./cmd/myapp
//	go tool toolgui-wasm serve ./cmd/myapp
//
// The app it builds needs a js/wasm entry point; see toolgui/tgwasm.
package main

import (
	"fmt"
	"log"
	"os"
)

const usage = `toolgui-wasm builds a ToolGUI app into a static site.

Usage:

	toolgui-wasm build [-o dir] [package]
	toolgui-wasm serve [-o dir] [-addr address] [package]

The package defaults to the current directory, and the output to ./dist.
`

func main() {
	log.SetFlags(0)
	log.SetPrefix("toolgui-wasm: ")

	if len(os.Args) < 2 {
		fmt.Fprint(os.Stderr, usage)
		os.Exit(2)
	}

	var err error
	switch os.Args[1] {
	case "build":
		err = runBuild(os.Args[2:])
	case "serve":
		err = runServe(os.Args[2:])
	case "help", "-h", "-help", "--help":
		fmt.Print(usage)
		return
	default:
		fmt.Fprintf(os.Stderr, "unknown command %q\n\n%s", os.Args[1], usage)
		os.Exit(2)
	}

	if err != nil {
		log.Fatal(err)
	}
}
