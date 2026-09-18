package main

import (
	"slices"
	"testing"

	"github.com/voilelab/toolgui/docs/demos"
	"github.com/voilelab/toolgui/toolgui/tgframe"
)

// Every demo page renders, the component pages included: an example that no
// longer compiles into a page -- an anchor that moved, a snippet that reads
// state the page does not have -- fails here rather than in the browser.
func TestEveryPageRenders(t *testing.T) {
	app := newApp()
	if err := addPluginDemo(app); err != nil {
		t.Fatalf("addPluginDemo: %v", err)
	}

	names := app.AppConf().PageNames
	if len(names) < 50 {
		t.Errorf("got %d pages, want one per component and then some",
			len(names))
	}

	for _, name := range names {
		if err := app.Run(name, tgframe.NewState(), func(tgframe.NotifyPack) {}); err != nil {
			t.Errorf("%s: %v", name, err)
		}
	}
}

// The component pages are what /demo/#/<component> lands on, so the book's
// links to them are only as good as the names being there.
func TestComponentPagesExist(t *testing.T) {
	app := newApp()

	have := map[string]bool{}
	for _, name := range app.AppConf().PageNames {
		have[name] = true
	}

	for _, name := range []string{
		"title", "metric", "dataframe", "chart", "slider", "form", "dialog",
		"iframe", "error", "color_picker",
	} {
		if !have[name] {
			t.Errorf("no page for %s", name)
		}
	}
}

// What the side nav lists: the category pages and the app's own, and none of
// the sixty odd component pages under them. Those are reached by url -- a
// link or an embed from the book -- and a list holding both reads as two
// lists spliced together.
func TestOnlyCategoryPagesAreListed(t *testing.T) {
	app := newApp()
	if err := addPluginDemo(app); err != nil {
		t.Fatalf("addPluginDemo: %v", err)
	}

	conf := app.AppConf()

	listed := []string{}
	for _, name := range conf.PageNames {
		if !conf.PageConfs[name].Hidden {
			listed = append(listed, name)
		}
	}

	want := []string{
		"index",
		"content", "data", "input", "layout", "misc",
		"sidebar", "app_menu", "function_cache", "code",
		// In no group, so its own page is the only way to it.
		"plugin",
	}

	if !slices.Equal(listed, want) {
		t.Errorf("the nav lists %v, want %v", listed, want)
	}
}

// The other half: every component's page is there, and hidden.
func TestComponentPagesAreHidden(t *testing.T) {
	conf := newApp().AppConf()

	for _, g := range demos.Groups() {
		for _, d := range g.Demos {
			pc, ok := conf.PageConfs[d.Name]
			if !ok {
				t.Errorf("no page for %s", d.Name)
				continue
			}

			if !pc.Hidden {
				t.Errorf("%s is listed in the nav", d.Name)
			}
		}
	}
}
