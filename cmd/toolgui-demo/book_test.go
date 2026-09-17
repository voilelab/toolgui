package main

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/voilelab/toolgui/docs/demos"
)

// demoDir and bookDir are docs/demos and docs/src, from this package.
var (
	demoDir = filepath.Join("..", "..", "docs", "demos")
	bookDir = filepath.Join("..", "..", "docs", "src")
)

// includeRe matches the mdBook directive the book slices a snippet with.
var includeRe = regexp.MustCompile(`\{\{#include ([^:{}]+):(\w+)\}\}`)

// anchorRe matches an opening anchor comment in a demo file.
var anchorRe = regexp.MustCompile(`ANCHOR: (\w+)`)

// TestBookIncludesResolve slices every snippet the book points at docs/demos.
// mdBook fails quietly -- an include it cannot resolve is left in the page as
// the directive it was written as -- so a renamed anchor is caught here
// rather than by a reader.
func TestBookIncludesResolve(t *testing.T) {
	found := 0

	for _, page := range bookPages(t) {
		for _, inc := range includesIn(t, page) {
			found++

			code, err := demos.Anchored(inc.file, inc.anchor)
			if err != nil {
				t.Errorf("%s: %v", page, err)
				continue
			}

			if strings.TrimSpace(code) == "" {
				t.Errorf("%s: %s:%s is empty", page, inc.file, inc.anchor)
			}
		}
	}

	if found == 0 {
		t.Fatal("the book includes nothing from docs/demos any more")
	}
}

// TestEveryAnchorIsShown catches the other direction: a snippet neither the
// book nor a demo page shows is a snippet nothing keeps honest.
func TestEveryAnchorIsShown(t *testing.T) {
	shown := map[string]bool{}

	for _, page := range bookPages(t) {
		for _, inc := range includesIn(t, page) {
			shown[inc.file+":"+inc.anchor] = true
		}
	}

	for _, d := range everyDemo() {
		for _, b := range d.Blocks {
			for _, a := range b.Anchors {
				shown[d.Name+".go:"+a] = true
			}
		}
	}

	files, err := filepath.Glob(filepath.Join(demoDir, "*.go"))
	if err != nil {
		t.Fatal(err)
	}

	for _, file := range files {
		bs, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}

		for _, m := range anchorRe.FindAllStringSubmatch(string(bs), -1) {
			if name := filepath.Base(file); !shown[name+":"+m[1]] {
				t.Errorf("%s: anchor %q is shown nowhere", name, m[1])
			}
		}
	}
}

// TestEveryBlockHasCode is the demo app's half: an example whose code column
// is empty means its anchor and its function have drifted apart.
func TestEveryBlockHasCode(t *testing.T) {
	for _, d := range everyDemo() {
		if len(d.Blocks) == 0 {
			t.Errorf("%s has no examples", d.Name)
		}

		for _, b := range d.Blocks {
			if strings.TrimSpace(b.Code) == "" {
				t.Errorf("%s: %s has no source", d.Name, b.ID)
			}
		}
	}
}

// everyDemo returns every component's demo, the plugin one included.
func everyDemo() []*demos.Demo {
	all := []*demos.Demo{demos.Plugin()}
	for _, g := range demos.Groups() {
		all = append(all, g.Demos...)
	}

	return all
}

// include is one {{#include}} pointed at docs/demos: the file it names, by
// base name, and the anchor it slices at.
type include struct {
	file   string
	anchor string
}

// includesIn returns the includes on page that reach into docs/demos. An
// include of anything else is some other page's business.
func includesIn(t *testing.T, page string) []include {
	t.Helper()

	bs, err := os.ReadFile(page)
	if err != nil {
		t.Fatal(err)
	}

	demoPath, err := filepath.Abs(demoDir)
	if err != nil {
		t.Fatal(err)
	}

	found := []include{}
	for _, m := range includeRe.FindAllStringSubmatch(string(bs), -1) {
		target, err := filepath.Abs(
			filepath.Join(filepath.Dir(page), filepath.FromSlash(m[1])))
		if err != nil {
			t.Fatal(err)
		}

		if filepath.Dir(target) != demoPath {
			continue
		}

		found = append(found, include{file: filepath.Base(target), anchor: m[2]})
	}

	return found
}

// bookPages returns every page of the book.
func bookPages(t *testing.T) []string {
	t.Helper()

	pages := []string{}
	err := filepath.WalkDir(bookDir,
		func(name string, _ os.DirEntry, err error) error {
			if err != nil {
				return err
			}

			if filepath.Ext(name) == ".md" {
				pages = append(pages, name)
			}

			return nil
		})
	if err != nil {
		t.Fatal(err)
	}

	if len(pages) == 0 {
		t.Fatal("no book pages found")
	}

	return pages
}
