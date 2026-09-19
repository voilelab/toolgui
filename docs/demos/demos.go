// Package demos holds the component examples the book and the demo app share.
//
// One file per component, one anchored snippet per example. The book slices a
// snippet out with mdBook's {{#include file.go:anchor}}, the demo app runs the
// function around it and prints the same slice beside what it drew. Neither
// side holds a copy, so an example cannot be right in one place and stale in
// the other.
//
// A snippet is written for the reader: it uses p.Main, as a page function
// does, and everything it needs that is not the example itself -- a helper, a
// type, the embed -- lives outside the anchors.
package demos

import (
	"embed"
	"fmt"
	"strings"

	"github.com/voilelab/toolgui/toolgui/tgframe"
)

// sources are the files the snippets are sliced out of, embedded so that the
// demo app carries them into a browser build, where there is no disk to read
// them from.
//
//go:embed *.go
var sources embed.FS

// Block is one Component / Code row: the example to run, and the source of it
// the book includes.
type Block struct {
	// ID names the column pair the row is drawn in, e.g. "show_title".
	ID string

	// Run writes the example into p.Main.
	Run tgframe.RunFunc

	// Anchors are the anchor comments Code was sliced at, which is what the
	// book names in its {{#include}}.
	Anchors []string

	// Code is the source between the anchors, dedented: the shared
	// indentation is the enclosing function's, and the code column is half a
	// page wide. The book includes the same slice with it, under prose that
	// has the whole page.
	Code string
}

// Demo is one component: a page of its own in the demo app, and the examples
// the book includes.
type Demo struct {
	// Name is the page name, and the base name of the file under docs/demos
	// the snippets come from.
	Name string

	// Title names the page in the side nav.
	Title string

	// Blocks are the rows, in the order the page shows them.
	Blocks []Block

	// Sidebar, when set, runs after the blocks: an example whose point is
	// that it is written into the side column.
	Sidebar tgframe.RunFunc
}

// Group is a coarse page, the way the demo was organised before every
// component had a page of its own.
type Group struct {
	Name  string
	Title string

	// LeadDivider draws a divider between the header row and the first
	// example, as the misc page has always had.
	LeadDivider bool

	// NumberDividers gives the dividers ids of their own -- "1", "2", ... --
	// as the input page has always had.
	NumberDividers bool

	Demos []*Demo
}

// snippet is one block before its source has been sliced out: the anchors
// name the slices, the enclosing demo names the file.
type snippet struct {
	id      string
	run     tgframe.RunFunc
	anchors []string
}

// show describes a row: the column id it is drawn in, the function to run,
// and the anchors its source is sliced at. More than one anchor is for an
// example the book walks through a piece at a time: the pieces are joined
// here, and included one by one there.
func show(id string, run tgframe.RunFunc, anchors ...string) snippet {
	return snippet{id: id, run: run, anchors: anchors}
}

// demo builds a component's demo, slicing each block's source out of
// <name>.go at the block's anchor.
func demo(name, title string, snippets ...snippet) *Demo {
	d := &Demo{Name: name, Title: title}

	for _, s := range snippets {
		parts := []string{}
		for _, a := range s.anchors {
			parts = append(parts, anchored(name+".go", a))
		}

		d.Blocks = append(d.Blocks, Block{
			ID:      s.id,
			Run:     s.run,
			Anchors: s.anchors,
			Code:    dedent(strings.Join(parts, "\n")),
		})
	}

	return d
}

// withSidebar hands the demo a side column example. It reads as part of the
// table demo calls are written in.
func (d *Demo) withSidebar(run tgframe.RunFunc) *Demo {
	d.Sidebar = run
	return d
}

// dedent drops the indentation every line of code shares, which for a
// snippet sliced out of a function body is that function's. A block's
// anchors are joined first, so the whole block is measured at once and its
// pieces keep the indentation they have relative to each other.
//
// A block holding a line that starts at column 0 -- the middle of a raw
// string, a top-level declaration -- shares nothing, and is left alone.
func dedent(code string) string {
	lines := strings.Split(code, "\n")

	shared := -1
	for _, l := range lines {
		if strings.TrimSpace(l) == "" {
			continue
		}

		indent := len(l) - len(strings.TrimLeft(l, "\t"))
		if shared == -1 || indent < shared {
			shared = indent
		}
	}

	if shared <= 0 {
		return code
	}

	out := make([]string, len(lines))
	for i, l := range lines {
		if strings.TrimSpace(l) == "" {
			out[i] = ""
			continue
		}

		out[i] = l[shared:]
	}

	return strings.Join(out, "\n")
}

// Anchored returns the lines of file between the opening and closing anchor
// comments name is written in. That is the slice mdBook's {{#include}} takes:
// the marker lines dropped, everything between them verbatim, joined by
// newlines with none at the end.
//
// It reads the embedded copy of the file, which is what the demo app ships,
// so it answers for the browser build as well as for a test with the
// repository on disk.
func Anchored(file, name string) (string, error) {
	bs, err := sources.ReadFile(file)
	if err != nil {
		return "", err
	}

	kept := []string{}
	found := false

	for _, line := range strings.Split(string(bs), "\n") {
		if marker(line, "ANCHOR_END:") == name {
			return strings.Join(kept, "\n"), nil
		}

		if marker(line, "ANCHOR:") == name {
			found = true
			continue
		}

		if found {
			kept = append(kept, strings.TrimSuffix(line, "\r"))
		}
	}

	return "", fmt.Errorf("%s: no closed anchor %q", file, name)
}

// anchored is Anchored where there is nothing to do about a miss: the table
// below names the anchors, so one that is not there is a snippet the book
// cannot include either, and the demo is not worth starting without it.
func anchored(file, name string) string {
	code, err := Anchored(file, name)
	if err != nil {
		panic("toolgui demos: " + err.Error())
	}

	return code
}

// marker returns the name a line gives after prefix, or "" for a line that is
// not one of mdBook's anchor comments.
func marker(line, prefix string) string {
	_, rest, ok := strings.Cut(line, prefix)
	if !ok {
		return ""
	}

	return strings.TrimSpace(rest)
}
