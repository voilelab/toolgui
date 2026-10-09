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

	toolgui-wasm build [-o dir] [-ldflags flags] [-manifest file] [-assets dir] [-lazy-assets dir] [-icon url] [-head file] [-offline] [package]
	toolgui-wasm serve [-o dir] [-ldflags flags] [-manifest file] [-assets dir] [-lazy-assets dir] [-icon url] [-head file] [-offline] [-addr address] [package]

The package defaults to the current directory, and the output to ./dist.
-ldflags is passed to go build as is. -manifest is written as manifest.json,
and -assets is copied to assets/, where the manifest's icons can point.
-icon replaces the favicon url in index.html, e.g. assets/favicon.svg.
-head is an html file inserted into the head of index.html, e.g. meta tags.
-offline writes a service worker, so the site opens with no network.
-lazy-assets is copied to assets/ too, but with -offline the worker keeps a
file only once the page fetches it, e.g. a large runtime the app may not use.
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
