// Package wasmweb holds the browser frontend: the page that loads a ToolGUI
// app compiled to wasm, and the worker that runs it.
//
// It is its own package so a server or desktop build never needs these files
// on disk; only cmd/toolgui-wasm, which writes them next to the app.wasm it
// builds, imports it.
//
// Build it with `task asset_wasm`.
package wasmweb

import (
	"embed"
	"io/fs"
)

//go:embed all:build
var assets embed.FS

// GetAssets return the frontend as a file system rooted at its index.html.
func GetAssets() fs.FS {
	fsys, err := fs.Sub(assets, "build")
	if err != nil {
		panic(err)
	}
	return fsys
}
