package tgframe

import (
	"errors"
	"io/fs"
	"net/http"
	"path"
	"regexp"

	"github.com/voilelab/toolgui/toolgui/tgutil"
)

// PluginAssetPrefix is the path plugin assets are served under. A file in the
// asset set registered as "gauge" is served at /plugin/gauge/<file>.
const PluginAssetPrefix = "/plugin/"

// ErrInvalidPluginName is the error that a plugin name is not usable in a url
// path.
var ErrInvalidPluginName = errors.New("invalid plugin name")

// ErrDuplicatedPluginName is the error that two asset sets are registered
// under one name.
var ErrDuplicatedPluginName = errors.New("duplicated plugin name")

// pluginNamePattern keeps the name to one url path segment, so it cannot
// escape the prefix it is served under.
var pluginNamePattern = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_-]*$`)

// AddPluginAssets serves fsys under /plugin/<name>/, so a plugin can ship as
// the files it is made of instead of a string in the page function.
//
//	//go:embed assets
//	var gaugeAssets embed.FS
//
//	sub, _ := fs.Sub(gaugeAssets, "assets")
//	app.AddPluginAssets("gauge", sub)
//
// The url of a file in the set is [PluginAssetURL].
func (app *App) AddPluginAssets(name string, fsys fs.FS) error {
	if !pluginNamePattern.MatchString(name) {
		return tgutil.Errorf("%w: `%s`", ErrInvalidPluginName, name)
	}

	if fsys == nil {
		return tgutil.NewError("nil plugin assets")
	}

	if _, exist := app.pluginAssets[name]; exist {
		return tgutil.Errorf("%w: `%s`", ErrDuplicatedPluginName, name)
	}

	app.pluginAssets[name] = fsys
	return nil
}

// PluginAssets return the registered asset sets by name. Executors serve them;
// a page function has no use for it.
func (app *App) PluginAssets() map[string]fs.FS {
	return app.pluginAssets
}

// PluginAssetURL return the url the given file of the named asset set is
// served at.
//
//	tgcomp.Plugin(p.Main, "gauge", tgframe.PluginAssetURL("gauge", "gauge.js"), props)
func PluginAssetURL(name, file string) string {
	return path.Join(PluginAssetPrefix, name, file)
}

// PluginAssetHandler serves an app's plugin assets under [PluginAssetPrefix].
// Executors mount it; it is what makes [PluginAssetURL] resolve.
//
// The set is looked up per request, so assets registered after the handler is
// built are still served.
func PluginAssetHandler(app *App) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET "+PluginAssetPrefix+"{name}/{file...}",
		func(resp http.ResponseWriter, req *http.Request) {
			name := req.PathValue("name")
			fsys, ok := app.PluginAssets()[name]
			if !ok {
				resp.WriteHeader(http.StatusNotFound)
				return
			}

			prefix := path.Join(PluginAssetPrefix, name)
			http.StripPrefix(prefix, http.FileServerFS(fsys)).ServeHTTP(resp, req)
		})

	return mux
}
