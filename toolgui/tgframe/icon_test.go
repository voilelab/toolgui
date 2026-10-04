package tgframe

import (
	"strings"
	"testing"

	"github.com/voilelab/toolgui/toolgui/tgjson"
)

func TestSetIcon(t *testing.T) {
	app := NewApp()

	// No icon: the key stays out, so the frontend falls back to page emoji.
	bs, err := tgjson.Marshal(app.AppConf())
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(bs), `"icon"`) {
		t.Errorf("AppConf has an icon key with no icon set: %s", bs)
	}

	app.SetIcon("assets/favicon.svg")
	if got := app.AppConf().Icon; got != "assets/favicon.svg" {
		t.Errorf("AppConf().Icon = %q, want assets/favicon.svg", got)
	}
}
