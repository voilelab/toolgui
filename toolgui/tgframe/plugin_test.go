package tgframe

import (
	"errors"
	"fmt"
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

// A url names something in the set. A file reaching upwards would otherwise
// be cleaned into a url outside the prefix, which the app does not serve.
func TestPluginAssetURLStaysInTheSet(t *testing.T) {
	cases := map[string]string{
		"/gauge.js":            "/plugin/gauge/gauge.js",
		"../../secret":         "/plugin/gauge/secret",
		"../other/other.js":    "/plugin/gauge/other/other.js",
		"nested/../nested.css": "/plugin/gauge/nested.css",
	}

	for file, want := range cases {
		if got := PluginAssetURL("gauge", file); got != want {
			t.Errorf("PluginAssetURL(gauge, %q) = %q, want %q", file, got, want)
		}
	}
}

// The map is the app's own bookkeeping: a caller that mutates what it reads
// back would be registering a set behind AddPluginAssets, name checks and all.
func TestPluginAssetsIsACopy(t *testing.T) {
	app := NewApp()
	if err := app.AddPluginAssets("gauge", testAssets()); err != nil {
		t.Fatalf("AddPluginAssets: %v", err)
	}

	sets := app.PluginAssets()
	delete(sets, "gauge")
	sets["../evil"] = testAssets()

	if _, ok := app.PluginAssets()["gauge"]; !ok {
		t.Error("gauge went missing from the app")
	}

	if _, ok := app.PluginAssets()["../evil"]; ok {
		t.Error("a set was registered without going through AddPluginAssets")
	}
}

// A set can be registered while the app is serving, so the two have to be
// safe against each other.
func TestPluginAssetsConcurrentRegisterAndServe(t *testing.T) {
	app := NewApp()
	handler := PluginAssetHandler(app)

	done := make(chan struct{})
	go func() {
		defer close(done)
		for i := range 50 {
			_ = app.AddPluginAssets(fmt.Sprintf("gauge%d", i), testAssets())
		}
	}()

	for range 50 {
		resp := httptest.NewRecorder()
		handler.ServeHTTP(resp,
			httptest.NewRequest(http.MethodGet, PluginAssetURL("gauge0", "gauge.js"), nil))
	}

	<-done
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
