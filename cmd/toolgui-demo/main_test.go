package main

import (
	"testing"

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
