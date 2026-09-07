package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindWasmExec(t *testing.T) {
	goroot := t.TempDir()

	_, err := findWasmExec(goroot)
	if err == nil {
		t.Fatal("expected an error for a GOROOT with no shim")
	}

	misc := filepath.Join(goroot, "misc", "wasm")
	writeFile(t, filepath.Join(misc, "wasm_exec.js"), "old")

	name, err := findWasmExec(goroot)
	if err != nil {
		t.Fatalf("findWasmExec: %v", err)
	}
	if name != filepath.Join(misc, "wasm_exec.js") {
		t.Errorf("got %s, want the misc/wasm copy", name)
	}

	// lib/wasm is where the toolchain keeps it now, so it wins.
	lib := filepath.Join(goroot, "lib", "wasm")
	writeFile(t, filepath.Join(lib, "wasm_exec.js"), "new")

	name, err = findWasmExec(goroot)
	if err != nil {
		t.Fatalf("findWasmExec: %v", err)
	}
	if name != filepath.Join(lib, "wasm_exec.js") {
		t.Errorf("got %s, want the lib/wasm copy", name)
	}
}

func TestWriteFrontend(t *testing.T) {
	out := t.TempDir()

	err := writeFrontend(out)
	if err != nil {
		t.Fatalf("writeFrontend: %v", err)
	}

	// The stub assets carry an index.html and nothing else, so that is all a
	// build without a yarn build can be checked for.
	_, err = os.Stat(filepath.Join(out, "index.html"))
	if err != nil {
		t.Errorf("no index.html in the site: %v", err)
	}
}

// TestBuild covers the whole command against a real app.
func TestBuild(t *testing.T) {
	if testing.Short() {
		t.Skip("compiles a wasm binary")
	}

	out := t.TempDir()

	err := build(out, "github.com/voilelab/toolgui/toolgui/tgwasm/example/hello")
	if err != nil {
		t.Fatalf("build: %v", err)
	}

	for _, name := range []string{"index.html", "app.wasm", "wasm_exec.js"} {
		info, err := os.Stat(filepath.Join(out, name))
		if err != nil {
			t.Errorf("no %s in the site: %v", name, err)
			continue
		}

		if info.Size() == 0 {
			t.Errorf("%s is empty", name)
		}
	}
}

func writeFile(t *testing.T, name, body string) {
	t.Helper()

	err := os.MkdirAll(filepath.Dir(name), 0o755)
	if err != nil {
		t.Fatal(err)
	}

	err = os.WriteFile(name, []byte(body), 0o644)
	if err != nil {
		t.Fatal(err)
	}
}
