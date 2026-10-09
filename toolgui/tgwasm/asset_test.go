//go:build js && wasm

package tgwasm

import (
	"errors"
	"syscall/js"
	"testing"
)

func TestAssetURL(t *testing.T) {
	cases := []struct {
		name  string
		base  any
		asset string
		want  string
		err   error
	}{
		{name: "root", base: "http://localhost:8080/", asset: "pyodide/pyodide.js",
			want: "http://localhost:8080/assets/pyodide/pyodide.js"},
		{name: "sub path", base: "https://u.github.io/repo/", asset: "a.txt",
			want: "https://u.github.io/repo/assets/a.txt"},
		{name: "index.html with query", base: "https://u.github.io/repo/index.html?embed#/p", asset: "a.txt",
			want: "https://u.github.io/repo/assets/a.txt"},
		{name: "escaped", base: "http://h/", asset: "d #1/a?b c.txt",
			want: "http://h/assets/d%20%231/a%3Fb%20c.txt"},
		{name: "dot dot", base: "http://h/", asset: "a/../../x", err: ErrBadAssetName},
		{name: "absolute", base: "http://h/", asset: "/etc/passwd", err: ErrBadAssetName},
		{name: "backslash", base: "http://h/", asset: `..\x`, err: ErrBadAssetName},
		{name: "empty", base: "http://h/", asset: "", err: ErrBadAssetName},
		{name: "unset", base: nil, asset: "a.txt", err: ErrNoBase},
		{name: "relative base", base: "repo/", asset: "a.txt", err: ErrNoBase},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if c.base == nil {
				js.Global().Delete(BaseName)
			} else {
				js.Global().Set(BaseName, c.base)
			}
			defer js.Global().Delete(BaseName)

			got, err := AssetURL(c.asset)
			if !errors.Is(err, c.err) {
				t.Fatalf("AssetURL(%q) error = %v, want %v", c.asset, err, c.err)
			}

			if got != c.want {
				t.Errorf("AssetURL(%q) = %q, want %q", c.asset, got, c.want)
			}
		})
	}
}
