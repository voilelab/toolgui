package tgexec

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/voilelab/toolgui/toolgui/tgframe"
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
	err = json.NewDecoder(resp.Body).Decode(&members)
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

	icons, _ := members["icons"].([]any)
	if len(icons) != 2 {
		t.Errorf("icons = %v, want 2 entries", members["icons"])
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
