package main

import (
	"io/fs"
	"os"
	"path/filepath"
	"slices"
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

// TestBuildDemoIsStillOneBinary builds the demo, which registers a page per
// component on top of the category pages it has always had. A page is not a
// binary: however many an app declares, a build writes the same files a one
// page app does, with the one app.wasm among them.
func TestBuildDemoIsStillOneBinary(t *testing.T) {
	if testing.Short() {
		t.Skip("compiles two wasm binaries")
	}

	hello := t.TempDir()
	err := build(hello, "github.com/voilelab/toolgui/toolgui/tgwasm/example/hello")
	if err != nil {
		t.Fatalf("build the one page app: %v", err)
	}

	demo := t.TempDir()
	err = build(demo, "github.com/voilelab/toolgui/cmd/toolgui-demo")
	if err != nil {
		t.Fatalf("build the demo: %v", err)
	}

	got, want := siteFiles(t, demo), siteFiles(t, hello)
	if !slices.Equal(got, want) {
		t.Errorf("the demo built %v, the one page app %v", got, want)
	}

	binaries := []string{}
	for _, name := range got {
		if filepath.Ext(name) == ".wasm" {
			binaries = append(binaries, name)
		}
	}

	if !slices.Equal(binaries, []string{"app.wasm"}) {
		t.Errorf("built %v, want the one app.wasm", binaries)
	}
}

// siteFiles returns what a build wrote under out, sorted, so two sites can be
// compared by what they are made of.
func siteFiles(t *testing.T, out string) []string {
	t.Helper()

	names := []string{}
	err := filepath.WalkDir(out, func(name string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if entry.IsDir() {
			return nil
		}

		rel, err := filepath.Rel(out, name)
		if err != nil {
			return err
		}

		names = append(names, filepath.ToSlash(rel))
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	slices.Sort(names)
	return names
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
