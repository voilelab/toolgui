package tgframe

import (
	"encoding/json"
	"slices"
	"strings"
	"testing"
)

func TestSetStickyQuery(t *testing.T) {
	app := NewApp()

	b, err := json.Marshal(app.AppConf())
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(b), "sticky_query") {
		t.Errorf("AppConf without SetStickyQuery has sticky_query: %s", b)
	}

	keys := []string{"group", "day"}
	app.SetStickyQuery(keys...)
	keys[0] = "changed"

	if got := app.AppConf().StickyQuery; !slices.Equal(got, []string{"group", "day"}) {
		t.Errorf("StickyQuery = %v, want [group day]", got)
	}

	app.SetStickyQuery()
	if got := app.AppConf().StickyQuery; len(got) != 0 {
		t.Errorf("StickyQuery = %v after SetStickyQuery(), want none", got)
	}
}
