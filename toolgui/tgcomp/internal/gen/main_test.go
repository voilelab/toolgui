package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

// TestGeneratedUpToDate fails when a tc* change was not followed by
// go generate in toolgui/tgcomp.
func TestGeneratedUpToDate(t *testing.T) {
	root := filepath.Join("..", "..")

	want, err := generate(root)
	if err != nil {
		t.Fatalf("generate: %v", err)
	}

	for name, src := range want {
		got, err := os.ReadFile(name)
		if err != nil {
			t.Errorf("%s: %v; run go generate in toolgui/tgcomp", name, err)
			continue
		}

		if !bytes.Equal(got, src) {
			t.Errorf("%s is stale; run go generate in toolgui/tgcomp", name)
		}
	}

	stale, err := filepath.Glob(filepath.Join(root, "*_gen.go"))
	if err != nil {
		t.Fatal(err)
	}

	for _, name := range stale {
		if _, ok := want[name]; !ok {
			t.Errorf("%s is no longer generated; remove it", name)
		}
	}
}

// TestConstNames checks every exported name of a marked const is forwarded,
// whether the declaration is grouped or binds several names on one line.
func TestConstNames(t *testing.T) {
	dir := t.TempDir()
	src := `package tcx

//tgcomp:export
const A, B, c = 1, 2, 3

//tgcomp:export
const (
	D = iota
	E
)
`
	if err := os.WriteFile(filepath.Join(dir, "x.go"), []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}

	out, err := generateTemp(t, dir)
	if err != nil {
		t.Fatalf("generatePackage: %v", err)
	}

	for _, want := range []string{"A = tcx.A", "B = tcx.B", "D = tcx.D", "E = tcx.E"} {
		if !bytes.Contains(out, []byte(want)) {
			t.Errorf("missing %q in:\n%s", want, out)
		}
	}

	if bytes.Contains(out, []byte("tcx.c")) {
		t.Errorf("unexported c forwarded:\n%s", out)
	}
}

// generateTemp generates a test package tcx from dir.
func generateTemp(t *testing.T, dir string) ([]byte, error) {
	t.Helper()

	names, err := markedNames(dir)
	if err != nil {
		t.Fatal(err)
	}

	return generatePackage(dir, "tcx", map[string]map[string]bool{"tcx": names})
}

// TestParamNames checks unnamed and blank params get names that don't clash
// with the ones already there.
func TestParamNames(t *testing.T) {
	dir := t.TempDir()
	src := `package tcx

//tgcomp:export
func F(_ string, p0 int) {}

//tgcomp:export
func G(int, string) {}
`
	if err := os.WriteFile(filepath.Join(dir, "x.go"), []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}

	out, err := generateTemp(t, dir)
	if err != nil {
		t.Fatalf("generatePackage: %v", err)
	}

	for _, want := range []string{
		"func F(p1 string, p0 int)", "tcx.F(p1, p0)",
		"func G(p0 int, p1 string)", "tcx.G(p0, p1)",
	} {
		if !bytes.Contains(out, []byte(want)) {
			t.Errorf("missing %q in:\n%s", want, out)
		}
	}
}

// TestDocQualifiers checks a doc comment's references to forwarded names
// point at tgcomp, and others are left alone.
func TestDocQualifiers(t *testing.T) {
	dir := t.TempDir()
	src := `package tcx

// F is like [tcx.G]; call tcx.F(c), not tcx.hidden or tcy.F.
//
//tgcomp:export
func F() {}

//tgcomp:export
func G() {}
`
	if err := os.WriteFile(filepath.Join(dir, "x.go"), []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}

	out, err := generateTemp(t, dir)
	if err != nil {
		t.Fatalf("generatePackage: %v", err)
	}

	want := "// F is like [G]; call tgcomp.F(c), not tcx.hidden or tcy.F."
	if !bytes.Contains(out, []byte(want)) {
		t.Errorf("missing %q in:\n%s", want, out)
	}
}

// TestDiscoversPackages checks a new tc* package with markers is generated
// without being listed anywhere, and one without markers is skipped.
func TestDiscoversPackages(t *testing.T) {
	root := t.TempDir()

	for name, src := range map[string]string{
		"tcnew":  "package tcnew\n\n//tgcomp:export\nfunc F() {}\n",
		"tcutil": "package tcutil\n\nfunc G() {}\n",
	} {
		if err := os.Mkdir(filepath.Join(root, name), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(root, name, "x.go"), []byte(src), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	out, err := generate(root)
	if err != nil {
		t.Fatalf("generate: %v", err)
	}

	if _, ok := out[filepath.Join(root, "new_gen.go")]; !ok || len(out) != 1 {
		t.Errorf("generated %v, want only new_gen.go", out)
	}
}
