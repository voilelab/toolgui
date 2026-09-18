package tgframe_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/voilelab/toolgui/toolgui/tgframe"
)

// A hidden page is a page in every way but the nav: it is in the conf, it is
// served, and only the frontend reading Hidden leaves it off the list.
func TestHiddenPage(t *testing.T) {
	app := tgframe.NewApp()
	app.AddPage("index", "Index", func(p *tgframe.Params) error { return nil })

	ran := false
	app.AddPageByConfig(&tgframe.PageConfig{
		Name:   "detail",
		Title:  "Detail",
		Hidden: true,
	}, func(p *tgframe.Params) error {
		ran = true
		return nil
	})

	conf := app.AppConf()
	if got, want := len(conf.PageNames), 2; got != want {
		t.Fatalf("got %d pages, want %d", got, want)
	}

	if conf.PageConfs["index"].Hidden {
		t.Error("index is hidden")
	}

	if !conf.PageConfs["detail"].Hidden {
		t.Error("detail is not hidden")
	}

	if err := app.Run("detail", tgframe.NewState(),
		func(tgframe.NotifyPack) {}); err != nil {
		t.Fatalf("run detail: %v", err)
	}

	if !ran {
		t.Error("the hidden page did not run")
	}
}

// The flag is absent for the pages that are not hidden, which is most of
// them: the frontend reads a missing key as a listed page.
func TestHiddenIsOmitted(t *testing.T) {
	bs, err := json.Marshal(&tgframe.PageConfig{Name: "index", Title: "Index"})
	if err != nil {
		t.Fatal(err)
	}

	if strings.Contains(string(bs), "hidden") {
		t.Errorf("a listed page marshals as %s", bs)
	}

	bs, err = json.Marshal(&tgframe.PageConfig{
		Name: "detail", Title: "Detail", Hidden: true})
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(bs), `"hidden":true`) {
		t.Errorf("a hidden page marshals as %s", bs)
	}
}
