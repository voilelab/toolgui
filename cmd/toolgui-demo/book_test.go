package main

import (
	"os"
	"path/filepath"
	"regexp"
	"slices"
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

// summaryRe matches a link to a component page in the book's contents, which
// is where the order the book reads in comes from.
var summaryRe = regexp.MustCompile(`\(components/(\w+)/(\w+)\.md\)`)

// TestDemoOrderFollowsTheBook holds the two lists to one order: the groups,
// and the components inside each one. A component the book has no page for is
// nobody's to place among the rest, so what is asked of it is only that it
// comes after them.
func TestDemoOrderFollowsTheBook(t *testing.T) {
	inBook := bookOrder(t)

	groups := []string{}
	for _, g := range demos.Groups() {
		groups = append(groups, g.Name)

		want := []string{}
		for _, name := range inBook[g.Name] {
			if hasDemo(g, name) {
				want = append(want, name)
			}
		}

		got := []string{}
		for _, d := range g.Demos {
			if slices.Contains(inBook[g.Name], d.Name) {
				got = append(got, d.Name)
			}
		}

		if !slices.Equal(got, want) {
			t.Errorf("%s: demos read %v, the book reads %v", g.Name, got, want)
		}

		// The contents place the rest against each other; one they leave out
		// has no place among them, so it goes after all of them rather than
		// between two the book put side by side.
		unlisted := ""
		for _, d := range g.Demos {
			if !slices.Contains(inBook[g.Name], d.Name) {
				unlisted = d.Name
				continue
			}

			if unlisted != "" {
				t.Errorf("%s: %s comes after %s, which the book does not list",
					g.Name, d.Name, unlisted)
			}
		}
	}

	if !slices.Equal(groups, inBook[""]) {
		t.Errorf("groups read %v, the book reads %v", groups, inBook[""])
	}
}

// bookOrder reads the contents: the groups under the key "", and each group's
// components under its own name, both in the order they are listed in.
func bookOrder(t *testing.T) map[string][]string {
	t.Helper()

	bs, err := os.ReadFile(filepath.Join(bookDir, "SUMMARY.md"))
	if err != nil {
		t.Fatal(err)
	}

	order := map[string][]string{}
	for _, m := range summaryRe.FindAllStringSubmatch(string(bs), -1) {
		group, page := m[1], m[2]

		// A group's own page is what names the group, in its place among the
		// others.
		if page == "index" {
			order[""] = append(order[""], group)
			continue
		}

		order[group] = append(order[group], page)
	}

	if len(order[""]) == 0 {
		t.Fatal("the contents list no component groups")
	}

	return order
}

// hasDemo reports whether the group has an example for the component, which
// not every page of the book does.
func hasDemo(g *demos.Group, name string) bool {
	return slices.ContainsFunc(g.Demos, func(d *demos.Demo) bool {
		return d.Name == name
	})
}
