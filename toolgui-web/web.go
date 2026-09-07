package toolguiweb

import (
	"embed"
	"io/fs"
	"path"
)

//go:embed app/build/index.html
var IndexBody string

//go:embed app/build/static/*
var staticDir embed.FS

//go:embed app/build/*
var rootAssets embed.FS

func GetStaticDir() fs.FS {
	fsys, err := fs.Sub(staticDir, path.Join("app", "build"))
	if err != nil {
		panic(err)
	}
	return fsys
}

func GetRootAssets() map[string][]byte {
	entries, err := rootAssets.ReadDir(path.Join("app", "build"))
	if err != nil {
		panic(err)
	}

	files := map[string][]byte{}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		bs, err := rootAssets.ReadFile(path.Join("app", "build", entry.Name()))
		if err != nil {
			panic(err)
		}

		files[entry.Name()] = bs
	}
	return files
}

// The browser build, which cmd/toolgui-wasm writes next to a user's app.wasm.
// It is a directory of files rather than one inlined bundle, and nothing
// serves it: a wasm app is static files on any host.
//
//go:embed all:wasm/build
var wasmDir embed.FS

// GetWasmDir return the browser frontend as a file system rooted at its
// index.html. Build it with `task asset_wasm`.
func GetWasmDir() fs.FS {
	fsys, err := fs.Sub(wasmDir, path.Join("wasm", "build"))
	if err != nil {
		panic(err)
	}
	return fsys
}
