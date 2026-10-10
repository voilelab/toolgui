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

	out, err := generatePackage(dir, "tcx")
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
