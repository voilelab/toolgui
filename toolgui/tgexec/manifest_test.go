package tgexec

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/voilelab/toolgui/toolgui/tgframe"
	"github.com/voilelab/toolgui/toolgui/tgjson"
)

// getManifest fetch /manifest.json and return the response and its members.
func getManifest(t *testing.T, url string) (*http.Response, map[string]any) {
	t.Helper()

	resp, err := http.Get(url + "/manifest.json")
	if err != nil {
		t.Fatalf("get manifest: %v", err)
	}
	t.Cleanup(func() { resp.Body.Close() })

	members := map[string]any{}
	err = tgjson.UnmarshalRead(resp.Body, &members)
	if err != nil {
		t.Fatalf("decode manifest: %v", err)
	}

	return resp, members
}

// An app that configures nothing still gets a usable manifest.
func TestManifestDefault(t *testing.T) {
	srv, _ := newTestServer(t)

	resp, members := getManifest(t, srv.URL)

	if got := resp.Header.Get("Content-Type"); got != "application/manifest+json" {
		t.Errorf("Content-Type = %q, want %q", got, "application/manifest+json")
	}

	if members["name"] != "ToolGUI App" {
		t.Errorf("name = %v, want %q", members["name"], "ToolGUI App")
	}

	if members["display"] != "standalone" {
		t.Errorf("display = %v, want %q", members["display"], "standalone")
	}

	// toolgui ships no icon, so the default advertises none.
	if _, ok := members["icons"]; ok {
		t.Errorf("icons = %v, want it left out", members["icons"])
	}
}

// The app title names the default manifest, so a titled app doesn't have to
// write one out to be named.
func TestManifestDefaultTakesAppTitle(t *testing.T) {
	app := tgframe.NewApp()
	app.SetTitle("My Tool")
	app.AddPage("index", "Index", func(p *tgframe.Params) error { return nil })

	e := NewWebExecutor(app)
	t.Cleanup(e.Destroy)

	mux, err := e.Mux()
	if err != nil {
		t.Fatalf("Mux: %v", err)
	}

	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	_, members := getManifest(t, srv.URL)

	if members["name"] != "My Tool" {
		t.Errorf("name = %v, want %q", members["name"], "My Tool")
	}

	if members["short_name"] != "My Tool" {
		t.Errorf("short_name = %v, want %q", members["short_name"], "My Tool")
	}

	// The rest of the default is still there.
	if members["display"] != "standalone" {
		t.Errorf("display = %v, want %q", members["display"], "standalone")
	}
}

// What the app sets is what the browser reads.
func TestManifestSet(t *testing.T) {
	srv, e := newTestServer(t)

	e.SetManifest(&Manifest{
		Name:      "My Tool",
		ShortName: "Tool",
		Icons: []ManifestIcon{
			{Src: "icon.png", Type: "image/png", Sizes: "512x512"},
		},
	})

	_, members := getManifest(t, srv.URL)

	if members["name"] != "My Tool" {
		t.Errorf("name = %v, want %q", members["name"], "My Tool")
	}

	if members["short_name"] != "Tool" {
		t.Errorf("short_name = %v, want %q", members["short_name"], "Tool")
	}

	// Unset members are left out rather than sent empty.
	if _, ok := members["display"]; ok {
		t.Errorf("display = %v, want it left out", members["display"])
	}
}

// A nil manifest is a reset, not a manifest with nothing in it.
func TestManifestSetNil(t *testing.T) {
	srv, e := newTestServer(t)

	e.SetManifest(&Manifest{Name: "My Tool"})
	e.SetManifest(nil)

	_, members := getManifest(t, srv.URL)

	if members["name"] != "ToolGUI App" {
		t.Errorf("name = %v, want %q", members["name"], "ToolGUI App")
	}
}

// Extra carries the members the struct doesn't name.
func TestManifestExtra(t *testing.T) {
	srv, e := newTestServer(t)

	e.SetManifest(&Manifest{
		Name: "My Tool",
		Extra: map[string]any{
			"categories": []string{"utilities"},

			// A named member the app would rather write itself.
			"display": "fullscreen",
		},
	})

	_, members := getManifest(t, srv.URL)

	categories, _ := members["categories"].([]any)
	if len(categories) != 1 || categories[0] != "utilities" {
		t.Errorf("categories = %v, want [utilities]", members["categories"])
	}

	if members["display"] != "fullscreen" {
		t.Errorf("display = %v, want %q", members["display"], "fullscreen")
	}

	if members["name"] != "My Tool" {
		t.Errorf("name = %v, want %q", members["name"], "My Tool")
	}
}

// Hash page name mode routes every path to an asset, but not the manifest.
func TestManifestInHashPageNameMode(t *testing.T) {
	app := tgframe.NewApp()
	app.SetHashPageNameMode(true)
	app.AddPage("index", "Index", func(p *tgframe.Params) error { return nil })

	e := NewWebExecutor(app)
	t.Cleanup(e.Destroy)

	mux, err := e.Mux()
	if err != nil {
		t.Fatalf("Mux: %v", err)
	}

	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	_, members := getManifest(t, srv.URL)
	if members["name"] != "ToolGUI App" {
		t.Errorf("name = %v, want %q", members["name"], "ToolGUI App")
	}
}

// A key in Extra wins over the named field of the same name, even when that
// field is set and would have been written. The merge is what the app reaches
// for to say something the struct cannot, so the struct must not win it back.
func TestManifestExtraWinsOverNamedField(t *testing.T) {
	bs, err := tgjson.Marshal(&Manifest{
		Name:    "My Tool",
		Display: "standalone",
		Extra: map[string]any{
			"display": "fullscreen",
		},
	})
	if err != nil {
		t.Fatalf("marshal manifest: %v", err)
	}

	members := map[string]any{}
	if err := tgjson.Unmarshal(bs, &members); err != nil {
		t.Fatalf("unmarshal manifest: %v", err)
	}

	if members["display"] != "fullscreen" {
		t.Errorf("display = %v, want %q", members["display"], "fullscreen")
	}

	// The members Extra says nothing about are still the struct's.
	if members["name"] != "My Tool" {
		t.Errorf("name = %v, want %q", members["name"], "My Tool")
	}
}

// A manifest with no Extra encodes straight from the struct, and a nested
// Manifest goes through MarshalJSONTo the same way a top-level one does.
func TestManifestMarshalsNested(t *testing.T) {
	bs, err := tgjson.Marshal(map[string]*Manifest{
		"manifest": {Name: "My Tool", Display: "standalone"},
	})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var out struct {
		Manifest struct {
			Name    string `json:"name"`
			Display string `json:"display"`
		} `json:"manifest"`
	}
	if err := tgjson.Unmarshal(bs, &out); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if out.Manifest.Name != "My Tool" || out.Manifest.Display != "standalone" {
		t.Errorf("manifest = %+v, want name and display set", out.Manifest)
	}
}
