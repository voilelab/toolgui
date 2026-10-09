package tgtest_test

import (
	"net/url"
	"slices"
	"strings"
	"testing"

	"github.com/voilelab/toolgui/toolgui/tgcomp"
	"github.com/voilelab/toolgui/toolgui/tgframe"
	"github.com/voilelab/toolgui/toolgui/tgtest"
)

func TestOpenWithQuery(t *testing.T) {
	app := newApp("detail", func(p *tgframe.Params) error {
		tgcomp.Text(p.Main, "group="+p.Query.Get("group"))
		return nil
	})

	p := tgtest.Open(t, app, "detail", tgtest.WithQuery(url.Values{
		"group": {"a"},
	}))
	if err := p.Err(); err != nil {
		t.Fatalf("Err = %v", err)
	}

	if !p.HasText("group=a") {
		t.Error("query not read")
	}
}

func TestQueryEmptyNotNil(t *testing.T) {
	app := newApp("index", func(p *tgframe.Params) error {
		if p.Query == nil {
			tgcomp.Text(p.Main, "nil")
		}
		return nil
	})

	p := tgtest.Open(t, app, "index")
	if p.HasText("nil") {
		t.Error("Query is nil without a query")
	}
}

// A page writing to its Query does not reach the next run.
func TestQueryCopiedPerRun(t *testing.T) {
	app := newApp("index", func(p *tgframe.Params) error {
		tgcomp.Text(p.Main, "x="+p.Query.Get("x"))
		p.Query.Set("x", "changed")
		return nil
	})

	p := tgtest.Open(t, app, "index", tgtest.WithQuery(url.Values{"x": {"1"}}))
	p.Rerun()

	if !p.HasText("x=1") {
		t.Error("a run saw the previous run's write")
	}
}

func TestPageLink(t *testing.T) {
	app := tgframe.NewApp()
	app.AddPage("index", "Index", func(p *tgframe.Params) error {
		tgcomp.PageLink(p.Main, "Open", "detail", url.Values{"group": {"a"}})
		return nil
	})
	app.AddPage("detail", "Detail", func(p *tgframe.Params) error {
		return nil
	})

	p := tgtest.Open(t, app, "index")
	if err := p.Err(); err != nil {
		t.Fatalf("Err = %v", err)
	}

	links := p.FindByName("page_link_component")
	if len(links) != 1 {
		t.Fatalf("got %d links, want 1", len(links))
	}

	link := links[0]
	if link.String("page") != "detail" || link.String("text") != "Open" {
		t.Errorf("props = %v", link.Props)
	}

	query, _ := link.Prop("query").(map[string]any)
	group, _ := query["group"].([]any)
	if len(group) != 1 || group[0] != "a" {
		t.Errorf("query = %v", link.Prop("query"))
	}
}

func TestPageLinkUnknownPage(t *testing.T) {
	app := newApp("index", func(p *tgframe.Params) error {
		tgcomp.PageLink(p.Main, "Open", "nowhere", nil)
		return nil
	})

	p := tgtest.Open(t, app, "index")
	if err := p.Err(); err == nil ||
		!strings.Contains(err.Error(), "nowhere") {
		t.Errorf("Err = %v, want page not found", err)
	}

	if len(p.FindByName("page_link_component")) != 0 {
		t.Error("an unknown page drew a link")
	}
}

func TestReplaceQuery(t *testing.T) {
	groups := []string{"a", "b", "c"}
	app := newApp("detail", func(p *tgframe.Params) error {
		conf := &tgcomp.SelectConf{}
		if def := slices.Index(groups, p.Query.Get("group")); def >= 0 {
			conf.SetDefault(def)
		}

		q := url.Values{}
		if idx := tgcomp.Select(p.Main, "Group", groups, conf); idx != nil {
			q.Set("group", groups[*idx])
		}
		p.ReplaceQuery(q)
		return nil
	})

	p := tgtest.Open(t, app, "detail", tgtest.WithQuery(url.Values{
		"group": {"b"},
	}))
	if got := p.Query().Get("group"); got != "b" {
		t.Errorf("Query before select = %q", got)
	}

	p.GetByLabel("Group").Select(2)
	if err := p.Err(); err != nil {
		t.Fatal(err)
	}
	if got := p.Query().Encode(); got != "group=c" {
		t.Errorf("Query = %q, want group=c", got)
	}
}

func TestNavigateFromDataFrame(t *testing.T) {
	ids := []string{"0003", "0004"}
	app := newApp("problems", func(p *tgframe.Params) error {
		sel := tgcomp.DataFrame(p.Main, []string{"id"},
			[][]string{{ids[0]}, {ids[1]}},
			&tgcomp.DataFrameConf{Selection: tgcomp.SelectionModeSingle})
		if len(sel) == 1 {
			p.Navigate("detail", url.Values{"id": {ids[sel[0]]}})
		}
		return nil
	})
	app.AddPage("detail", "Detail", func(*tgframe.Params) error { return nil })

	p := tgtest.Open(t, app, "problems")
	if _, _, ok := p.Navigated(); ok {
		t.Error("navigated before a row was picked")
	}

	p.FindByName("dataframe_component")[0].Select(1)
	if err := p.Err(); err != nil {
		t.Fatal(err)
	}

	page, q, ok := p.Navigated()
	if !ok || page != "detail" || q.Encode() != "id=0004" {
		t.Errorf("Navigated = %q %q %v", page, q.Encode(), ok)
	}
}
