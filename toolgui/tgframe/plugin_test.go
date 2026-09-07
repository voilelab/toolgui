package tgframe

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"
)

func testAssets() fstest.MapFS {
	return fstest.MapFS{
		"gauge.js": &fstest.MapFile{Data: []byte("window.toolgui.update(1)")},
	}
}

func TestAddPluginAssets(t *testing.T) {
	app := NewApp()

	if err := app.AddPluginAssets("gauge", testAssets()); err != nil {
		t.Fatalf("AddPluginAssets: %v", err)
	}

	err := app.AddPluginAssets("gauge", testAssets())
	if !errors.Is(err, ErrDuplicatedPluginName) {
		t.Errorf("second AddPluginAssets = %v, want ErrDuplicatedPluginName", err)
	}
}

// A name is one url path segment, so it cannot reach outside the prefix its
// assets are served under.
func TestAddPluginAssetsRejectsNameThatIsNotOneSegment(t *testing.T) {
	for _, name := range []string{"", "..", "a/b", "a/../b", "a b", "/gauge"} {
		err := NewApp().AddPluginAssets(name, testAssets())
		if !errors.Is(err, ErrInvalidPluginName) {
			t.Errorf("AddPluginAssets(%q) = %v, want ErrInvalidPluginName", name, err)
		}
	}
}

func TestPluginAssetURL(t *testing.T) {
	if got := PluginAssetURL("gauge", "gauge.js"); got != "/plugin/gauge/gauge.js" {
		t.Errorf("PluginAssetURL = %q, want /plugin/gauge/gauge.js", got)
	}
}

func TestPluginAssetHandler(t *testing.T) {
	app := NewApp()
	if err := app.AddPluginAssets("gauge", testAssets()); err != nil {
		t.Fatalf("AddPluginAssets: %v", err)
	}

	handler := PluginAssetHandler(app)

	resp := httptest.NewRecorder()
	handler.ServeHTTP(resp,
		httptest.NewRequest(http.MethodGet, PluginAssetURL("gauge", "gauge.js"), nil))

	if resp.Code != http.StatusOK {
		t.Fatalf("code = %d, want 200", resp.Code)
	}

	if resp.Body.String() != "window.toolgui.update(1)" {
		t.Errorf("body = %q", resp.Body.String())
	}
}

// The set is read per request, so a plugin registered after an executor built
// its handler is still served.
func TestPluginAssetHandlerReadsTheSetPerRequest(t *testing.T) {
	app := NewApp()
	handler := PluginAssetHandler(app)

	if err := app.AddPluginAssets("gauge", testAssets()); err != nil {
		t.Fatalf("AddPluginAssets: %v", err)
	}

	resp := httptest.NewRecorder()
	handler.ServeHTTP(resp,
		httptest.NewRequest(http.MethodGet, PluginAssetURL("gauge", "gauge.js"), nil))

	if resp.Code != http.StatusOK {
		t.Errorf("code = %d, want 200", resp.Code)
	}
}

func TestPluginAssetHandlerUnknownName(t *testing.T) {
	handler := PluginAssetHandler(NewApp())

	resp := httptest.NewRecorder()
	handler.ServeHTTP(resp,
		httptest.NewRequest(http.MethodGet, PluginAssetURL("gauge", "gauge.js"), nil))

	if resp.Code != http.StatusNotFound {
		t.Errorf("code = %d, want 404", resp.Code)
	}
}
