package main

import (
	"testing"

	"github.com/voilelab/toolgui/toolgui/tgframe"
)

// Every demo page renders: Echo parses the lambda it is handed, so a page
// whose source moved around fails here rather than in the browser.
func TestEveryPageRenders(t *testing.T) {
	app := newApp()
	if err := addPluginDemo(app); err != nil {
		t.Fatalf("addPluginDemo: %v", err)
	}

	for _, name := range []string{
		"index", "content", "data", "input", "layout", "misc",
		"sidebar", "function_cache", "code", "plugin",
	} {
		if err := app.Run(name, tgframe.NewState(), func(tgframe.NotifyPack) {}); err != nil {
			t.Errorf("%s: %v", name, err)
		}
	}
}
