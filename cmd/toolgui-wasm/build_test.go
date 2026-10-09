package main

import (
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/voilelab/toolgui/toolgui/tgexec"
	"github.com/voilelab/toolgui/toolgui/tgjson"
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

// The default here and the one WebExecutor serves are the same manifest.
func TestDefaultManifestMatchesTgexec(t *testing.T) {
	bs, err := tgjson.Marshal(tgexec.DefaultManifest())
	if err != nil {
		t.Fatal(err)
	}

	want := map[string]any{}
	err = tgjson.Unmarshal(bs, &want)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(defaultManifest, want) {
		t.Errorf("defaultManifest = %v, tgexec serves %v", defaultManifest, want)
	}
}

func readManifest(t *testing.T, out string) map[string]any {
	t.Helper()

	bs, err := os.ReadFile(filepath.Join(out, "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}

	members := map[string]any{}
	err = tgjson.Unmarshal(bs, &members)
	if err != nil {
		t.Fatal(err)
	}

	return members
}

func TestWriteManifestDefault(t *testing.T) {
	out := t.TempDir()

	err := writeManifest(out, "")
	if err != nil {
		t.Fatalf("writeManifest: %v", err)
	}

	if got := readManifest(t, out)["name"]; got != "ToolGUI App" {
		t.Errorf("name = %v, want the default", got)
	}
}

// A manifest.json already in the web root is someone's, not ours to replace.
func TestWriteManifestKeepsExisting(t *testing.T) {
	out := t.TempDir()
	writeFile(t, filepath.Join(out, "manifest.json"), `{"name":"Mine"}`)

	err := writeManifest(out, "")
	if err != nil {
		t.Fatalf("writeManifest: %v", err)
	}

	if got := readManifest(t, out)["name"]; got != "Mine" {
		t.Errorf("name = %v, want the existing one kept", got)
	}
}

func TestWriteManifestFromFile(t *testing.T) {
	out := t.TempDir()
	writeFile(t, filepath.Join(out, "manifest.json"), `{"name":"Old"}`)

	src := filepath.Join(t.TempDir(), "app.json")
	writeFile(t, src, `{"name":"My Tool","icons":[{"src":"assets/icon.png"}]}`)

	err := writeManifest(out, src)
	if err != nil {
		t.Fatalf("writeManifest: %v", err)
	}

	if got := readManifest(t, out)["name"]; got != "My Tool" {
		t.Errorf("name = %v, want the one from -manifest", got)
	}
}

func TestWriteManifestRejectsBadJSON(t *testing.T) {
	for _, body := range []string{`{"name":`, `null`, `[]`, `"x"`} {
		src := filepath.Join(t.TempDir(), "app.json")
		writeFile(t, src, body)

		err := writeManifest(t.TempDir(), src)
		if err == nil {
			t.Errorf("expected an error for manifest %s", body)
		}
	}
}

// Copying assets into a directory inside them would never end.
func TestWriteAssetsRejectsOutInside(t *testing.T) {
	src := t.TempDir()
	writeFile(t, filepath.Join(src, "icon.png"), "png")

	for _, out := range []string{src, filepath.Join(src, "dist")} {
		err := writeAssets(src, out)
		if err == nil {
			t.Errorf("expected an error for output %s", out)
		}
	}

	// A sibling is fine.
	out := filepath.Join(t.TempDir(), "dist")
	err := writeAssets(src, out)
	if err != nil {
		t.Fatalf("writeAssets: %v", err)
	}

	_, err = os.Stat(filepath.Join(out, "assets", "icon.png"))
	if err != nil {
		t.Errorf("icon not copied: %v", err)
	}
}

func TestWriteFSOverwrites(t *testing.T) {
	src := t.TempDir()
	writeFile(t, filepath.Join(src, "icons", "icon.png"), "new")

	out := filepath.Join(t.TempDir(), "assets")
	writeFile(t, filepath.Join(out, "icons", "icon.png"), "old")

	err := writeFS(os.DirFS(src), out)
	if err != nil {
		t.Fatalf("writeFS: %v", err)
	}

	bs, err := os.ReadFile(filepath.Join(out, "icons", "icon.png"))
	if err != nil {
		t.Fatal(err)
	}
	if string(bs) != "new" {
		t.Errorf("icon = %q, want it overwritten", bs)
	}
}

// TestBuild covers the whole command against a real app.
func TestBuild(t *testing.T) {
	if testing.Short() {
		t.Skip("compiles a wasm binary")
	}

	out := t.TempDir()
	lazy := t.TempDir()
	writeFile(t, filepath.Join(lazy, "runtime.wasm"), "runtime")

	err := build(buildOpts{out: out, pkg: "github.com/voilelab/toolgui/toolgui/tgwasm/example/hello", lazy: lazy})
	if err != nil {
		t.Fatalf("build: %v", err)
	}

	// Without -offline, -lazy-assets is a plain copy.
	_, err = os.Stat(filepath.Join(out, "sw.js"))
	if err == nil {
		t.Error("sw.js written without -offline")
	}

	for _, name := range []string{"index.html", "manifest.json", "app.wasm", "wasm_exec.js", "assets/runtime.wasm"} {
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
	err := build(buildOpts{out: hello, pkg: "github.com/voilelab/toolgui/toolgui/tgwasm/example/hello"})
	if err != nil {
		t.Fatalf("build the one page app: %v", err)
	}

	demo := t.TempDir()
	err = build(buildOpts{out: demo, pkg: "github.com/voilelab/toolgui/cmd/toolgui-demo"})
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

func TestMarkIndex(t *testing.T) {
	name := filepath.Join(t.TempDir(), "index.html")
	writeFile(t, name, "<html><head>\n  <title>x</title>\n  </head><body></body></html>")

	// Twice: a rebuild into the same directory must not add a second one.
	for range 2 {
		err := markIndex(name)
		if err != nil {
			t.Fatalf("markIndex: %v", err)
		}
	}

	bs, err := os.ReadFile(name)
	if err != nil {
		t.Fatal(err)
	}

	if n := strings.Count(string(bs), offlineMeta); n != 1 {
		t.Errorf("%d metas in %s, want 1", n, bs)
	}

	if strings.Index(string(bs), offlineMeta) > strings.Index(string(bs), "</head>") {
		t.Errorf("meta outside the head: %s", bs)
	}
}

func TestWriteHead(t *testing.T) {
	dir := t.TempDir()
	index := filepath.Join(dir, "index.html")
	src := filepath.Join(dir, "head.html")
	writeFile(t, index, "<html><head>\n  <title>x</title>\n  </head><body></body></html>")
	writeFile(t, src, "<meta name=\"a\" />\n<meta name=\"b\" />\n")

	// Twice: a rebuild into the same directory must not add a second one.
	for range 2 {
		err := writeHead(index, src)
		if err != nil {
			t.Fatalf("writeHead: %v", err)
		}
	}

	bs, err := os.ReadFile(index)
	if err != nil {
		t.Fatal(err)
	}

	html := string(bs)
	snippet := "<meta name=\"a\" />\n<meta name=\"b\" />"
	if n := strings.Count(html, snippet); n != 1 {
		t.Errorf("%d snippets in %s, want 1", n, bs)
	}

	// Ahead of the head's links, so a CSP meta covers them.
	if strings.Index(html, snippet) > strings.Index(html, "<title>") {
		t.Errorf("snippet not at the start of the head: %s", bs)
	}
}

func TestWriteHeadMissingFile(t *testing.T) {
	dir := t.TempDir()
	index := filepath.Join(dir, "index.html")
	writeFile(t, index, "<html><head></head></html>")

	err := writeHead(index, filepath.Join(dir, "nope.html"))
	if err == nil {
		t.Error("writeHead with a missing file: no error")
	}
}

func TestWriteServiceWorker(t *testing.T) {
	out := t.TempDir()

	err := writeFrontend(out)
	if err != nil {
		t.Fatal(err)
	}

	// The stub frontend has no head to mark.
	writeFile(t, filepath.Join(out, "index.html"), "<html><head></head></html>")
	writeFile(t, filepath.Join(out, "app.wasm"), "wasm")
	writeFile(t, filepath.Join(out, "wasm_exec.js"), "shim")
	writeFile(t, filepath.Join(out, "manifest.json"), "{}")

	assets := t.TempDir()
	writeFile(t, filepath.Join(assets, "icons", "icon.png"), "png")
	writeFile(t, filepath.Join(assets, "a#b?c d.png"), "png")

	err = writeAssets(assets, out)
	if err != nil {
		t.Fatal(err)
	}

	lazy := t.TempDir()
	writeFile(t, filepath.Join(lazy, "pyodide", "pyodide.wasm"), "py")

	err = writeAssets(lazy, out)
	if err != nil {
		t.Fatal(err)
	}

	sw := func() string {
		t.Helper()

		err := writeServiceWorker(out, assets, lazy)
		if err != nil {
			t.Fatalf("writeServiceWorker: %v", err)
		}

		bs, err := os.ReadFile(filepath.Join(out, "sw.js"))
		if err != nil {
			t.Fatal(err)
		}

		return string(bs)
	}

	first := sw()
	for _, name := range []string{`"index.html"`, `"app.wasm"`, `"wasm_exec.js"`, `"manifest.json"`, `"assets/icons/icon.png"`} {
		if !strings.Contains(first, name) {
			t.Errorf("sw.js does not cache %s", name)
		}
	}

	if !strings.Contains(first, `"assets/a%23b%3Fc%20d.png"`) {
		t.Error("sw.js does not escape a # or ? in a file name")
	}

	// Lazy files are listed apart, so install does not fetch them.
	files, lazyList, _ := strings.Cut(first, "const LAZY = ")
	if strings.Contains(files, "pyodide") {
		t.Error("sw.js precaches a lazy asset")
	}
	if !strings.HasPrefix(lazyList, `["assets/pyodide/pyodide.wasm"]`) {
		t.Errorf("sw.js lazy list = %.60q", lazyList)
	}

	if strings.Contains(first, "__VERSION__") || strings.Contains(first, "__FILES__") || strings.Contains(first, "__LAZY__") {
		t.Error("sw.js still has a placeholder")
	}

	if sw() != first {
		t.Error("the same files gave another sw.js")
	}

	// A new binary is a new version, so browsers install it.
	writeFile(t, filepath.Join(out, "app.wasm"), "wasm 2")
	second := sw()
	if second == first {
		t.Error("a changed app.wasm kept the same sw.js")
	}

	// So is a new lazy file, which drops the old copy.
	writeFile(t, filepath.Join(out, "assets", "pyodide", "pyodide.wasm"), "py 2")
	if sw() == second {
		t.Error("a changed lazy asset kept the same sw.js")
	}
}

func TestWriteServiceWorkerNoLazy(t *testing.T) {
	out := t.TempDir()
	writeFile(t, filepath.Join(out, "index.html"), "<html><head></head></html>")
	writeFile(t, filepath.Join(out, "app.wasm"), "wasm")
	writeFile(t, filepath.Join(out, "wasm_exec.js"), "shim")
	writeFile(t, filepath.Join(out, "manifest.json"), "{}")

	err := writeServiceWorker(out, "", "")
	if err != nil {
		t.Fatalf("writeServiceWorker: %v", err)
	}

	bs, err := os.ReadFile(filepath.Join(out, "sw.js"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(bs), "const LAZY = []") {
		t.Error("sw.js without -lazy-assets has no empty LAZY")
	}
}

func TestCheckAssetConflict(t *testing.T) {
	assets := t.TempDir()
	writeFile(t, filepath.Join(assets, "icon.png"), "png")
	writeFile(t, filepath.Join(assets, "lib", "a.js"), "a")

	for _, name := range []string{
		"icon.png",   // same file
		"lib/a.js",   // same file, nested
		"icon.png/x", // a file used as a directory
		"lib",        // a directory used as a file
	} {
		lazy := t.TempDir()
		writeFile(t, filepath.Join(lazy, filepath.FromSlash(name)), "x")

		err := checkAssetConflict(assets, lazy)
		if err == nil {
			t.Errorf("no error for lazy %s", name)
		}
	}

	lazy := t.TempDir()
	writeFile(t, filepath.Join(lazy, "lib", "b.js"), "b")
	writeFile(t, filepath.Join(lazy, "pyodide", "pyodide.wasm"), "py")

	err := checkAssetConflict(assets, lazy)
	if err != nil {
		t.Errorf("checkAssetConflict: %v", err)
	}
}

func TestWriteIcon(t *testing.T) {
	src, err := os.ReadFile(filepath.Join("..", "..", "toolgui-web", "wasm", "index.html"))
	if err != nil {
		t.Fatal(err)
	}

	name := filepath.Join(t.TempDir(), "index.html")
	writeFile(t, name, string(src))

	err = writeIcon(name, "assets/fav&icon.svg")
	if err != nil {
		t.Fatalf("writeIcon: %v", err)
	}

	bs, err := os.ReadFile(name)
	if err != nil {
		t.Fatal(err)
	}

	page := string(bs)
	if n := strings.Count(page, `rel="icon"`); n != 1 {
		t.Errorf("%d icon links, want the one rewritten", n)
	}
	if !strings.Contains(page, `href="assets/fav&amp;icon.svg"`) {
		t.Errorf("icon href not rewritten: %s", page)
	}
	if strings.Contains(page, "data:image/svg+xml") {
		t.Error("the emoji icon is still there")
	}
}

func TestWriteIconNoLink(t *testing.T) {
	name := filepath.Join(t.TempDir(), "index.html")
	writeFile(t, name, "<html><head></head></html>")

	if err := writeIcon(name, "assets/favicon.svg"); err == nil {
		t.Error("expected an error for an index.html with no icon link")
	}
}
