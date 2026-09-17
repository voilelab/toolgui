package main

import (
	"os"
	"regexp"
	"testing"

	"github.com/voilelab/toolgui/docs/demos"
)

// embedRe matches the marker a component page puts where its live demo goes.
// docs/demo-embed.js turns it into an iframe onto the demo app's page for the
// component it names.
var embedRe = regexp.MustCompile(`data-toolgui-demo="(\w+)"`)

// noEmbed are the demos no page of the book embeds, and why. Everything else
// the demo app can show in a browser belongs on the page of the component it
// is an example of.
var noEmbed = map[string]string{
	"plugin": "a plugin is loaded over a url, and only a server serves one",
	"error":  "no component page of its own -- it is an architecture chapter",
}

// TestEveryEmbedIsADemo holds the marker to a page the browser build has. A
// name with no page behind it is an iframe onto the demo app's "page not
// found", which nothing else would catch.
func TestEveryEmbedIsADemo(t *testing.T) {
	inBrowser := map[string]bool{}
	for _, g := range demos.Groups() {
		for _, d := range g.Demos {
			inBrowser[d.Name] = true
		}
	}

	for _, page := range bookPages(t) {
		for _, name := range embedsIn(t, page) {
			if !inBrowser[name] {
				t.Errorf("%s: no demo named %q in the browser build", page, name)
			}
		}
	}
}

// TestOneEmbedPerPage is the one wasm instance a page is allowed: every demo
// in the frame is one worker with a copy of the binary in it, and a component
// page in the demo app already shows every example the component has.
func TestOneEmbedPerPage(t *testing.T) {
	for _, page := range bookPages(t) {
		if names := embedsIn(t, page); len(names) > 1 {
			t.Errorf("%s: embeds %v; one demo to a page", page, names)
		}
	}
}

// TestEveryDemoIsEmbedded is the other direction: a component whose page
// shows its code but not the demo behind it is a page a reader cannot try.
func TestEveryDemoIsEmbedded(t *testing.T) {
	embedded := map[string]bool{}
	for _, page := range bookPages(t) {
		for _, name := range embedsIn(t, page) {
			embedded[name] = true
		}
	}

	for _, d := range everyDemo() {
		if why, ok := noEmbed[d.Name]; ok {
			if embedded[d.Name] {
				t.Errorf("%s is embedded, but %s", d.Name, why)
			}

			continue
		}

		if !embedded[d.Name] {
			t.Errorf("%s: no page embeds it", d.Name)
		}
	}
}

// TestEmbedMatchesItsPage catches the pair drifting: the demo a page embeds
// is the demo whose code it includes, or the reader is reading one component
// and operating another.
func TestEmbedMatchesItsPage(t *testing.T) {
	for _, page := range bookPages(t) {
		included := map[string]bool{}
		for _, inc := range includesIn(t, page) {
			included[inc.file] = true
		}

		for _, name := range embedsIn(t, page) {
			if !included[name+".go"] {
				t.Errorf("%s: embeds %s, includes no code from %s.go",
					page, name, name)
			}
		}
	}
}

// embedsIn returns the demos page marks a live embed for, in the order it
// writes them.
func embedsIn(t *testing.T, page string) []string {
	t.Helper()

	bs, err := os.ReadFile(page)
	if err != nil {
		t.Fatal(err)
	}

	names := []string{}
	for _, m := range embedRe.FindAllStringSubmatch(string(bs), -1) {
		names = append(names, m[1])
	}

	return names
}
