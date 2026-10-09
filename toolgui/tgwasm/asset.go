//go:build js && wasm

package tgwasm

import (
	"net/url"
	"strings"
	"syscall/js"

	"github.com/voilelab/toolgui/toolgui/tgutil"
)

// BaseName is the global the page sets the app's root URL on, before the wasm
// program starts.
const BaseName = "toolguiBase"

// ErrNoBase is returned by [AssetURL] when the page set no root URL.
var ErrNoBase = tgutil.NewError("no app root url, the page did not set " + BaseName)

// ErrBadAssetName is returned by [AssetURL] for a name that is empty,
// absolute, or steps out of assets/.
var ErrBadAssetName = tgutil.NewError("bad asset name")

// AssetURL returns the absolute URL of name in the directory toolgui-wasm
// -assets copied, such as `pyodide/pyodide.js`. It is resolved against the
// directory of the app's index.html, not the worker's own location, so it
// holds wherever the app is hosted and in a frame.
//
// Each segment of name is escaped, so a `#`, `?` or space stays part of the
// path. Absolute names and `..` are refused.
func AssetURL(name string) (string, error) {
	segs := strings.Split(name, "/")
	for i, seg := range segs {
		if seg == "" || seg == "." || seg == ".." || strings.Contains(seg, `\`) {
			return "", tgutil.Errorf("%w: %q", ErrBadAssetName, name)
		}

		segs[i] = url.PathEscape(seg)
	}

	base, err := appBase()
	if err != nil {
		return "", err
	}

	return base + "assets/" + strings.Join(segs, "/"), nil
}

// appBase is the app's root URL from the page, ending in a slash.
func appBase() (string, error) {
	v := js.Global().Get(BaseName)
	if v.Type() != js.TypeString || v.String() == "" {
		return "", ErrNoBase
	}

	u, err := url.Parse(v.String())
	if err != nil || !u.IsAbs() {
		return "", tgutil.Errorf("%w: %q", ErrNoBase, v.String())
	}

	// The directory, should the page hand over index.html itself.
	u.RawQuery = ""
	u.Fragment = ""
	s := u.String()
	return s[:strings.LastIndex(s, "/")+1], nil
}
