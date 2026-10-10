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
